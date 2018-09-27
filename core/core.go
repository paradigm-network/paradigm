package core

import (
	"crypto/ecdsa"
	"fmt"
	"sort"
	"time"

	"github.com/paradigm-network/paradigm/common/crypto"
	"github.com/paradigm-network/paradigm/core/sequentia"
	"github.com/paradigm-network/paradigm/types"
	"github.com/paradigm-network/paradigm/storage"
	"github.com/paradigm-network/paradigm/common/log"
	"github.com/rs/zerolog"
)

type Core struct {
	id     int
	key    *ecdsa.PrivateKey
	pubKey []byte
	hexID  string
	cg     *sequentia.CometGraph

	participants        map[string]int //[PubKey] => id
	reverseParticipants map[int]string //[id] => PubKey
	Head                string
	Seq                 int

	transactionPool    [][]byte
	blockSignaturePool []types.BlockSignature

	logger *zerolog.Logger
}

func NewCore(
	id int,
	key *ecdsa.PrivateKey,
	participants map[string]int,
	store storage.Store,
	commitCh chan types.Block,
	) Core {

	reverseParticipants := make(map[int]string)
	for pk, id := range participants {
		reverseParticipants[id] = pk
	}

	core := Core{
		id:                  id,
		key:                 key,
		cg:                  sequentia.BuildCometGraph(participants, store, commitCh),
		participants:        participants,
		reverseParticipants: reverseParticipants,
		transactionPool:     [][]byte{},
		blockSignaturePool:  []types.BlockSignature{},
		logger:              log.GetLogger("Core"),
	}
	return core
}

func (c *Core) ID() int {
	return c.id
}

func (c *Core) PubKey() []byte {
	if c.pubKey == nil {
		c.pubKey = crypto.FromECDSAPub(&c.key.PublicKey)
	}
	return c.pubKey
}

func (c *Core) HexID() string {
	if c.hexID == "" {
		pubKey := c.PubKey()
		c.hexID = fmt.Sprintf("0x%X", pubKey)
	}
	return c.hexID
}

func (c *Core) Init() error {
	//Create and save the first Event
	initialEvent := types.NewComet([][]byte(nil), nil,
		[]string{"", ""},
		c.PubKey(),
		c.Seq)
	//We want to make the initial Event deterministic so that when a node is
	//restarted it will initialize the same Event. cf. github issues 19 and 10
	initialEvent.Body.Timestamp = time.Time{}.UTC()
	err := c.SignAndInsertSelfEvent(initialEvent)

	c.logger.Debug().
		Int("index",initialEvent.Index()).
		Str("hash",initialEvent.Hex()).
		Msg("Initial Event")
	return err
}

func (c *Core) Bootstrap() error {
	if err := c.cg.Bootstrap(); err != nil {
		return err
	}

	var head string
	var seq int

	last, isRoot, err := c.cg.Store.LastEventFrom(c.HexID())
	if err != nil {
		return err
	}

	if isRoot {
		root, err := c.cg.Store.GetRoot(c.HexID())
		if err != nil {
			head = root.X
			seq = root.Index
		}
	} else {
		lastEvent, err := c.GetComet(last)
		if err != nil {
			return err
		}
		head = last
		seq = lastEvent.Index()
	}

	c.Head = head
	c.Seq = seq

	return nil
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

func (c *Core) SignAndInsertSelfEvent(event types.Comet) error {
	if err := event.Sign(c.key); err != nil {
		return err
	}
	if err := c.InsertEvent(event, true); err != nil {
		return err
	}
	return nil
}

func (c *Core) InsertEvent(event types.Comet, setWireInfo bool) error {
	if err := c.cg.InsertComet(event, setWireInfo); err != nil {
		return err
	}
	if event.Creator() == c.HexID() {
		c.Head = event.Hex()
		c.Seq = event.Index()
	}
	return nil
}

func (c *Core) KnownEvents() map[int]int {
	return c.cg.KnownEvents()
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

func (c *Core) SignBlock(block types.Block) (types.BlockSignature, error) {
	sig, err := block.Sign(c.key)
	if err != nil {
		return types.BlockSignature{}, err
	}
	if err := block.SetSignature(sig); err != nil {
		return types.BlockSignature{}, err
	}
	return sig, c.cg.Store.SetBlock(block)
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

func (c *Core) OverSyncLimit(knownEvents map[int]int, syncLimit int) bool {
	totUnknown := 0
	myKnownEvents := c.KnownEvents()
	for i, li := range myKnownEvents {
		if li > knownEvents[i] {
			totUnknown += li - knownEvents[i]
		}
	}
	if totUnknown > syncLimit {
		return true
	}
	return false
}

func (c *Core) GetFrame() (types.Frame, error) {
	return c.cg.GetFrame()
}

//returns events that c knowns about and are not in 'known'
func (c *Core) EventDiff(known map[int]int) (events []types.Comet, err error) {
	unknown := []types.Comet{}
	//known represents the indez of the last event known for every participant
	//compare this to our view of events and fill unknown with events that we know of
	// and the other doesnt
	for id, ct := range known {
		pk := c.reverseParticipants[id]
		//get participant Events with index > ct
		participantEvents, err := c.cg.Store.ParticipantEvents(pk, ct)
		if err != nil {
			return []types.Comet{}, err
		}
		for _, e := range participantEvents {
			ev, err := c.cg.Store.GetComet(e)
			if err != nil {
				return []types.Comet{}, err
			}
			unknown = append(unknown, ev)
		}
	}
	sort.Sort(types.ByTopologicalOrder(unknown))

	return unknown, nil
}

func (c *Core) Sync(unknownEvents []types.WireEvent) error {

	c.logger.Debug().Int("unknown_events",len(unknownEvents)).
		Int("transaction_pool",len(c.transactionPool)).
		Int("block_signature_pool",len(c.blockSignaturePool)).Msg("Sync")
	otherHead := ""
	//add unknown events
	for k, we := range unknownEvents {
		ev, err := c.cg.ReadWireInfo(we)
		if err != nil {
			return err
		}
		if err := c.InsertEvent(*ev, false); err != nil {
			return err
		}
		//assume last event corresponds to other-head
		if k == len(unknownEvents)-1 {
			otherHead = ev.Hex()
		}
	}

	//create new event with self head and other head
	//only if there are pending loaded events or the pools are not empty
	if len(unknownEvents) > 0 ||
		len(c.transactionPool) > 0 ||
		len(c.blockSignaturePool) > 0 {

		newHead := types.NewComet(c.transactionPool, c.blockSignaturePool,
			[]string{c.Head, otherHead},
			c.PubKey(),
			c.Seq+1)

		if err := c.SignAndInsertSelfEvent(newHead); err != nil {
			return fmt.Errorf("Error inserting new head: %s", err)
		}

		//empty the pools
		c.transactionPool = [][]byte{}
		c.blockSignaturePool = []types.BlockSignature{}
	}

	return nil
}

func (c *Core) AddSelfEvent() error {
	if len(c.transactionPool) == 0 && len(c.blockSignaturePool) == 0 {
		c.logger.Debug().Msg("Empty transaction pool and block signature pool")
		return nil
	}

	//create new event with self head and empty other parent
	//empty transaction pool in its payload
	newHead := types.NewComet(c.transactionPool,
		c.blockSignaturePool,
		[]string{c.Head, ""},
		c.PubKey(), c.Seq+1)

	if err := c.SignAndInsertSelfEvent(newHead); err != nil {
		return fmt.Errorf("Error inserting new head: %s", err)
	}

	c.logger.Debug().
		Int("transactions",len(c.transactionPool)).
		Int("block_signatures",len(c.blockSignaturePool)).
		Msg("Created Self-Event")
	c.transactionPool = [][]byte{}
	c.blockSignaturePool = []types.BlockSignature{}

	return nil
}

func (c *Core) FromWire(wireEvents []types.WireEvent) ([]types.Comet, error) {
	events := make([]types.Comet, len(wireEvents), len(wireEvents))
	for i, w := range wireEvents {
		ev, err := c.cg.ReadWireInfo(w)
		if err != nil {
			return nil, err
		}
		events[i] = *ev
	}
	return events, nil
}

func (c *Core) ToWire(events []types.Comet) ([]types.WireEvent, error) {
	wireEvents := make([]types.WireEvent, len(events), len(events))
	for i, e := range events {
		wireEvents[i] = e.ToWire()
	}
	return wireEvents, nil
}

func (c *Core) RunConsensus() error {
	start := time.Now()
	err := c.cg.DivideRounds()
	c.logger.Debug().Int64("duration",time.Since(start).Nanoseconds()).Msg("DivideRounds()")
	if err != nil {
		c.logger.Error().Err(err).Msg("DivideRounds")
		return err
	}

	start = time.Now()
	err = c.cg.DecideFame()
	c.logger.Debug().Int64("duration",time.Since(start).Nanoseconds()).Msg("DecideFame()")
	if err != nil {
		c.logger.Error().Err(err).Msg("DecideFame")
		return err
	}

	start = time.Now()
	err = c.cg.FindOrder()
	c.logger.Debug().Int64("duration",time.Since(start).Nanoseconds()).Msg("FindOrder()")
	if err != nil {
		c.logger.Error().Err(err).Msg("FindOrder")
		return err
	}

	return nil
}

func (c *Core) AddTransactions(txs [][]byte) {
	c.transactionPool = append(c.transactionPool, txs...)
}

func (c *Core) AddBlockSignature(bs types.BlockSignature) {
	c.blockSignaturePool = append(c.blockSignaturePool, bs)
}

func (c *Core) GetHead() (types.Comet, error) {
	return c.cg.Store.GetComet(c.Head)
}

func (c *Core) GetComet(hash string) (types.Comet, error) {
	return c.cg.Store.GetComet(hash)
}

func (c *Core) GetCometTransactions(hash string) ([][]byte, error) {
	var txs [][]byte
	ex, err := c.GetComet(hash)
	if err != nil {
		return txs, err
	}
	txs = ex.Transactions()
	return txs, nil
}

func (c *Core) GetConsensusEvents() []string {
	return c.cg.ConsensusEvents()
}

func (c *Core) GetConsensusEventsCount() int {
	return c.cg.Store.ConsensusEventsCount()
}

func (c *Core) GetUndeterminedEvents() []string {
	return c.cg.UndeterminedEvents
}

func (c *Core) GetPendingLoadedEvents() int {
	return c.cg.PendingLoadedEvents
}

func (c *Core) GetConsensusTransactions() ([][]byte, error) {
	txs := [][]byte{}
	for _, e := range c.GetConsensusEvents() {
		eTxs, err := c.GetCometTransactions(e)
		if err != nil {
			return txs, fmt.Errorf("Consensus event not found: %s", e)
		}
		txs = append(txs, eTxs...)
	}
	return txs, nil
}

func (c *Core) GetLastConsensusRoundIndex() *int {
	return c.cg.LastConsensusRound
}

func (c *Core) GetConsensusTransactionsCount() int {
	return c.cg.ConsensusTransactions
}

func (c *Core) GetLastCommitedRoundEventsCount() int {
	return c.cg.LastCommitedRoundEvents
}

func (c *Core) GetLastBlockIndex() int {
	return c.cg.LastBlockIndex
}

func (c *Core) NeedGossip() bool {
	return c.cg.PendingLoadedEvents > 0 ||
		len(c.transactionPool) > 0 ||
		len(c.blockSignaturePool) > 0
}
