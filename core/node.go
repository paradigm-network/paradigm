package core

import (
	"sync"
	"time"
	"crypto/ecdsa"
	"fmt"
	"strconv"
	"github.com/paradigm-network/paradigm/core/sequentia"
	"github.com/paradigm-network/paradigm/network"
	"github.com/paradigm-network/paradigm/types"
	"github.com/paradigm-network/paradigm/common/timer"
	"github.com/paradigm-network/paradigm/storage"
	"github.com/paradigm-network/paradigm/network/peer"
	"github.com/paradigm-network/paradigm/proxy"
	"github.com/rs/zerolog"
	"github.com/paradigm-network/paradigm/common/log"
)

type Node struct {
	nodeState

	conf   *Config
	logger *zerolog.Logger

	id       int
	core     *Core
	coreLock sync.Mutex

	localAddr string

	peerSelector sequentia.PeerSelector
	selectorLock sync.Mutex

	trans network.Transport

	netCh <-chan network.RPC
	proxy proxy.AppProxy

	submitCh chan []byte

	commitCh chan types.Block

	shutdownCh chan struct{}

	controlTimer *timer.ControlTimer

	start        time.Time
	syncRequests int
	syncErrors   int
}

func NewNode(conf *Config,
	id int,
	key *ecdsa.PrivateKey,
	participants []peer.Peer,
	store storage.Store,
	trans network.Transport,
	proxy proxy.AppProxy,
) *Node {

	localAddr := trans.LocalAddr()

	pmap, _ := store.Participants()

	commitCh := make(chan types.Block, 400)
	core := NewCore(id, key, pmap, store, commitCh)

	peerSelector := sequentia.NewRandomPeerSelector(participants, localAddr)

	node := Node{
		id:           id,
		conf:         conf,
		core:         &core,
		localAddr:    localAddr,
		logger:       log.GetLogger("Node(" + string(id) + ")"),
		peerSelector: peerSelector,
		trans:        trans,
		netCh:        trans.Consumer(),
		proxy:        proxy,
		submitCh:     proxy.SubmitCh(),
		commitCh:     commitCh,
		shutdownCh:   make(chan struct{}),
		controlTimer: timer.NewRandomControlTimer(conf.HeartbeatTimeout),
	}

	//Initialize as Booting
	node.setStarting(true)
	node.setState(Booting)

	return &node
}

func (n *Node) Init(bootstrap bool) error {
	peerAddresses := []string{}
	for _, p := range n.peerSelector.Peers() {
		peerAddresses = append(peerAddresses, p.NetAddr)
	}
	n.logger.Info().Strs("peers", peerAddresses).Msg("Init Node")

	if bootstrap {
		n.logger.Info().Msg("Bootstrap")
		return n.core.Bootstrap()
	}
	return n.core.Init()
}

func (n *Node) RunAsync(gossip bool) {
	n.logger.Info().Msg("Run async")
	go n.Run(gossip)
}

func (n *Node) Run(gossip bool) {
	//The ControlTimer allows the background routines to control the
	//heartbeat timer when the node is in the Booting state. The timer should
	//only be running when there are uncommitted transactions in the system.
	go n.controlTimer.Run()

	//Execute some background work regardless of the state of the node.
	//Process RPC requests as well as SumbitTx and CommitBlock requests
	n.goFunc(n.doBackgroundWork)

	//Execute Node State Machine
	for {
		// Run different routines depending on node state
		state := n.getState()
		n.logger.Info().Str("state", state.String()).Msg("Run loop")

		switch state {
		case Booting:
			n.startGossipTimer(gossip)
		case CatchingUp:
			n.fastForward()
		case Shutdown:
			return
		}
	}
}

func (n *Node) doBackgroundWork() {
	for {
		select {
		case rpc := <-n.netCh:
			n.logger.Info().Msg("Processing RPC")
			n.processRPC(rpc)
			if n.core.NeedGossip() && !n.controlTimer.Set {
				n.controlTimer.ResetCh <- struct{}{}
			}
		case t := <-n.submitCh:
			n.logger.Info().Msg("Adding Transaction")
			n.addTransaction(t)
			if !n.controlTimer.Set {
				n.controlTimer.ResetCh <- struct{}{}
			}
		case block := <-n.commitCh:
			n.logger.Info().
				Int("index", block.Index()).
				Int("round_received", block.RoundReceived()).
				Int("txs", len(block.Transactions())).Msg("Committing Block")
			if err := n.commit(block); err != nil {
				n.logger.Error().Err(err).Msg("Committing Block")
			}
		case <-n.shutdownCh:
			return
		}
	}
}

func (n *Node) startGossipTimer(gossip bool) {
	for {
		oldState := n.getState()
		select {
		case <-n.controlTimer.TickCh:
			if gossip {
				proceed, err := n.preGossip()
				if proceed && err == nil {
					n.logger.Info().Msg("Time to gossip!")
					peer := n.peerSelector.Next()
					n.goFunc(func() { n.gossip(peer.NetAddr) })
				}
			}
			if !n.core.NeedGossip() {
				n.logger.Info().Msg("Gossip controlTimer stopped, because NeedGossip=false!")
				n.controlTimer.StopCh <- struct{}{}
			} else if !n.controlTimer.Set {
				n.controlTimer.ResetCh <- struct{}{}
			}
		case <-n.shutdownCh:
			return
		}

		newState := n.getState()
		if newState != oldState {
			return
		}
	}
}

func (n *Node) processRPC(rpc network.RPC) {

	if s := n.getState(); s != Booting {
		n.logger.Info().Str("state", s.String()).Msg("Discarding RPC Request")
		//XXX Use a SyncResponse by default but this should be either a special
		//ErrorResponse type or a type that corresponds to the request
		resp := &network.SyncResponse{
			FromID: n.id,
		}
		rpc.Respond(resp, fmt.Errorf("not ready: %s", s.String()))
		return
	}

	switch cmd := rpc.Command.(type) {
	case *network.SyncRequest:
		n.processSyncRequest(rpc, cmd)
	case *network.EagerSyncRequest:
		n.processEagerSyncRequest(rpc, cmd)
	default:
		n.logger.Info().
			Interface("cmd", rpc.Command).
			Msg("Unexpected RPC command")
		rpc.Respond(nil, fmt.Errorf("unexpected command"))
	}
}

func (n *Node) processSyncRequest(rpc network.RPC, cmd *network.SyncRequest) {
	n.logger.Info().
		Int("from_id", cmd.FromID).
		Interface("known", cmd.Known).
		Msg("Process SyncRequest")
	resp := &network.SyncResponse{
		FromID: n.id,
	}
	var respErr error

	//Check sync limit
	n.coreLock.Lock()
	overSyncLimit := n.core.OverSyncLimit(cmd.Known, n.conf.SyncLimit)
	n.coreLock.Unlock()
	if overSyncLimit {
		n.logger.Info().Msg("SyncLimit")
		resp.SyncLimit = true
	} else {
		//Compute Diff
		start := time.Now()
		n.coreLock.Lock()
		eventDiff, err := n.core.EventDiff(cmd.Known)
		n.coreLock.Unlock()

		elapsed := time.Since(start)
		n.logger.Info().Int64("duration", elapsed.Nanoseconds()).Msg("Diff()")
		if err != nil {
			n.logger.Error().Err(err).Msg("Calculating Diff")
			respErr = err
		}

		//Convert to WireEvents
		wireEvents, err := n.core.ToWire(eventDiff)
		if err != nil {
			n.logger.Error().Err(err).Msg("Converting to WireEvent")
			respErr = err
		} else {
			resp.Events = wireEvents
		}
	}

	//Get Self Known
	n.coreLock.Lock()
	knownEvents := n.core.KnownEvents()
	n.coreLock.Unlock()
	resp.Known = knownEvents

	n.logger.Info().
		Int("events", len(resp.Events)).
		Interface("known", resp.Known).
		Bool("sync_limit", resp.SyncLimit).
		Err(respErr).
		Msg("Responding to SyncRequest")
	rpc.Respond(resp, respErr)
}

func (n *Node) processEagerSyncRequest(rpc network.RPC, cmd *network.EagerSyncRequest) {
	n.logger.Info().
		Int("from_id", cmd.FromID).
		Int("events", len(cmd.Events)).
		Msg("EagerSyncRequest")

	success := true
	n.coreLock.Lock()
	err := n.sync(cmd.Events)
	n.coreLock.Unlock()
	if err != nil {
		n.logger.Error().Err(err).Msg("sync()")
		success = false
	}

	resp := &network.EagerSyncResponse{
		FromID:  n.id,
		Success: success,
	}
	rpc.Respond(resp, err)
}

func (n *Node) preGossip() (bool, error) {
	n.coreLock.Lock()
	defer n.coreLock.Unlock()

	//Check if it is necessary to gossip
	needGossip := n.core.NeedGossip() || n.isStarting()
	if !needGossip {
		n.logger.Info().Msg("Nothing to gossip")
		return false, nil
	}

	//If the transaction pool is not empty, create a new self-event and empty the
	//transaction pool in its payload
	if err := n.core.AddSelfEvent(); err != nil {
		n.logger.Error().Err(err).Msg("Adding SelfEvent")
		return false, err
	}

	return true, nil
}

func (n *Node) gossip(peerAddr string) error {
	//pull
	syncLimit, otherKnownEvents, err := n.pull(peerAddr)
	if err != nil {
		return err
	}

	//check and handle syncLimit
	if syncLimit {
		n.logger.Info().Str("from", peerAddr).Msg("SyncLimit")
		n.setState(CatchingUp)
		return nil
	}

	//push
	err = n.push(peerAddr, otherKnownEvents)
	if err != nil {
		return err
	}

	//update peer selector
	n.selectorLock.Lock()
	n.peerSelector.UpdateLast(peerAddr)
	n.selectorLock.Unlock()

	n.logStats()

	n.setStarting(false)

	return nil
}

func (n *Node) pull(peerAddr string) (syncLimit bool, otherKnownEvents map[int]int, err error) {
	//Compute Known
	n.coreLock.Lock()
	knownEvents := n.core.KnownEvents()
	n.logger.Info().
		Int("my_id", n.id).
		Interface("my_known", knownEvents).
		Msg("GetLocalKnownEvents:KnownEvents()")
	n.coreLock.Unlock()

	//Send SyncRequest
	start := time.Now()
	resp, err := n.requestSync(peerAddr, knownEvents)
	elapsed := time.Since(start)
	n.logger.Info().Int64("duration", elapsed.Nanoseconds()).Msg("requestSync()")
	if err != nil {
		n.logger.Error().Err(err).Msg("requestSync()")
		return false, nil, err
	}
	n.logger.Info().
		Int("from_id", resp.FromID).
		Bool("sync_limit", resp.SyncLimit).
		Int("events", len(resp.Events)).
		Interface("known", resp.Known).
		Msg("SyncResponse")

	if resp.SyncLimit {
		return true, nil, nil
	}

	//Add Comets to Sequentia and create new Head if necessary
	n.coreLock.Lock()
	err = n.sync(resp.Events)
	n.coreLock.Unlock()
	if err != nil {
		n.logger.Error().Err(err).Msg("sync()")
		return false, nil, err
	}

	return false, resp.Known, nil
}

func (n *Node) push(peerAddr string, knownEvents map[int]int) error {

	//Check SyncLimit
	n.coreLock.Lock()
	overSyncLimit := n.core.OverSyncLimit(knownEvents, n.conf.SyncLimit)
	n.coreLock.Unlock()
	if overSyncLimit {
		n.logger.Info().Msg("SyncLimit")
		return nil
	}

	//Compute Diff
	start := time.Now()
	n.coreLock.Lock()
	eventDiff, err := n.core.EventDiff(knownEvents)
	n.coreLock.Unlock()
	elapsed := time.Since(start)
	n.logger.Info().Int64("duration", elapsed.Nanoseconds()).Msg("Diff()")
	if err != nil {
		n.logger.Error().Err(err).Msg("Calculating Diff")
		return err
	}

	//Convert to WireEvents
	wireEvents, err := n.core.ToWire(eventDiff)
	if err != nil {
		n.logger.Error().Err(err).Msg("Converting to WireEvent")
		return err
	}

	//Create and Send EagerSyncRequest
	start = time.Now()
	resp2, err := n.requestEagerSync(peerAddr, wireEvents)
	elapsed = time.Since(start)
	n.logger.Info().Int64("duration", elapsed.Nanoseconds()).Interface("request_wireEvent", wireEvents).Interface("response", resp2).Msg("requestEagerSync()")
	if err != nil {
		n.logger.Error().Err(err).Msg("requestEagerSync()")
		return err
	}
	n.logger.Info().
		Int("from_id", resp2.FromID).
		Bool("success", resp2.Success).
		Msg("EagerSyncResponse")

	return nil
}

func (n *Node) fastForward() error {
	n.logger.Info().Msg("IN CATCHING-UP STATE")
	n.logger.Info().Msg("fast-sync not implemented yet")

	//XXX Work in Progress on fsync branch

	n.setState(Booting)

	return nil
}

func (n *Node) requestSync(target string, known map[int]int) (network.SyncResponse, error) {

	args := network.SyncRequest{
		FromID: n.id,
		Known:  known,
	}

	var out network.SyncResponse
	err := n.trans.Sync(target, &args, &out)

	return out, err
}

func (n *Node) requestEagerSync(target string, events []types.WireEvent) (network.EagerSyncResponse, error) {
	args := network.EagerSyncRequest{
		FromID: n.id,
		Events: events,
	}

	var out network.EagerSyncResponse
	err := n.trans.EagerSync(target, &args, &out)

	return out, err
}

func (n *Node) sync(events []types.WireEvent) error {
	//Insert Comets in Paradigm and create new Head if necessary
	start := time.Now()
	err := n.core.Sync(events)
	elapsed := time.Since(start)
	n.logger.Info().Int64("duration", elapsed.Nanoseconds()).Msg("Processed Sync()")
	if err != nil {
		return err
	}

	//Run consensus methods
	start = time.Now()
	err = n.core.RunConsensus()
	elapsed = time.Since(start)
	n.logger.Info().Int64("duration", elapsed.Nanoseconds()).Msg("Processed RunConsensus()")
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) commit(block types.Block) error {

	stateHash, err := n.proxy.CommitBlock(block)
	n.logger.Info().
		Int("block", block.Index()).
		Str("state_hash", fmt.Sprintf("0x%X", stateHash)).
		Err(err).
		Msg("CommitBlock Response")

	block.Body.StateHash = stateHash

	n.coreLock.Lock()
	defer n.coreLock.Unlock()
	sig, err := n.core.SignBlock(block)
	if err != nil {
		return err
	}
	n.core.AddBlockSignature(sig)

	return err
}

func (n *Node) addTransaction(tx []byte) {
	n.coreLock.Lock()
	defer n.coreLock.Unlock()
	n.core.AddTransactions([][]byte{tx})
}

func (n *Node) Shutdown() {
	if n.getState() != Shutdown {
		n.logger.Info().Msg("Shutdown")

		//Exit any non-shutdown state immediately
		n.setState(Shutdown)

		//Stop and wait for concurrent operations
		close(n.shutdownCh)
		n.waitRoutines()

		//For some reason this needs to be called after closing the shutdownCh
		//Not entirely sure why...
		n.controlTimer.Shutdown()

		//transport and store should only be closed once all concurrent operations
		//are finished otherwise they will panic trying to use close objects
		n.trans.Close()
		n.core.cg.Store.Close()
	}
}

func (n *Node) GetStats() map[string]string {
	toString := func(i *int) string {
		if i == nil {
			return "nil"
		}
		return strconv.Itoa(*i)
	}

	timeElapsed := time.Since(n.start)

	consensusEvents := n.core.GetConsensusEventsCount()
	consensusEventsPerSecond := float64(consensusEvents) / timeElapsed.Seconds()

	lastConsensusRound := n.core.GetLastConsensusRoundIndex()
	var consensusRoundsPerSecond float64
	if lastConsensusRound != nil {
		consensusRoundsPerSecond = float64(*lastConsensusRound) / timeElapsed.Seconds()
	}

	s := map[string]string{
		"last_consensus_round":   toString(lastConsensusRound),
		"last_block_index":       strconv.Itoa(n.core.GetLastBlockIndex()),
		"consensus_events":       strconv.Itoa(consensusEvents),
		"consensus_transactions": strconv.Itoa(n.core.GetConsensusTransactionsCount()),
		"undetermined_events":    strconv.Itoa(len(n.core.GetUndeterminedEvents())),
		"transaction_pool":       strconv.Itoa(len(n.core.transactionPool)),
		"num_peers":              strconv.Itoa(len(n.peerSelector.Peers())),
		"sync_rate":              strconv.FormatFloat(n.SyncRate(), 'f', 2, 64),
		"events_per_second":      strconv.FormatFloat(consensusEventsPerSecond, 'f', 2, 64),
		"rounds_per_second":      strconv.FormatFloat(consensusRoundsPerSecond, 'f', 2, 64),
		"round_events":           strconv.Itoa(n.core.GetLastCommitedRoundEventsCount()),
		"id":                     strconv.Itoa(n.id),
		"state":                  n.getState().String(),
	}
	return s
}

func (n *Node) logStats() {
	stats := n.GetStats()
	n.logger.Info().Interface("stat", stats)
}

func (n *Node) SyncRate() float64 {
	var syncErrorRate float64
	if n.syncRequests != 0 {
		syncErrorRate = float64(n.syncErrors) / float64(n.syncRequests)
	}
	return 1 - syncErrorRate
}

func (n *Node) GetBlock(blockIndex int) (types.Block, error) {
	return n.core.cg.Store.GetBlock(blockIndex)
}
