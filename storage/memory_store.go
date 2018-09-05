package storage

import (
	"strconv"

	"github.com/paradigm-network/paradigm/common"
	"github.com/paradigm-network/paradigm/errors"
	"github.com/paradigm-network/paradigm/types"
)

type InmemStore struct {
	cacheSize              int
	participants           map[string]int
	eventCache             *common.LRU
	roundCache             *common.LRU
	blockCache             *common.LRU
	consensusCache         *common.RollingIndex
	totConsensusEvents     int
	participantEventsCache *ParticipantEventsCache
	roots                  map[string]types.Root
	lastRound              int
}

func NewInmemStore(participants map[string]int, cacheSize int) *InmemStore {
	roots := make(map[string]types.Root)
	for pk := range participants {
		roots[pk] = types.NewBaseRoot()
	}
	return &InmemStore{
		cacheSize:              cacheSize,
		participants:           participants,
		eventCache:             common.NewLRU(cacheSize, nil),
		roundCache:             common.NewLRU(cacheSize, nil),
		blockCache:             common.NewLRU(cacheSize, nil),
		consensusCache:         common.NewRollingIndex(cacheSize),
		participantEventsCache: NewParticipantEventsCache(cacheSize, participants),
		roots:     roots,
		lastRound: -1,
	}
}

func (s *InmemStore) CacheSize() int {
	return s.cacheSize
}

func (s *InmemStore) Participants() (map[string]int, error) {
	return s.participants, nil
}

func (s *InmemStore) GetComet(key string) (types.Comet, error) {
	res, ok := s.eventCache.Get(key)
	if !ok {
		return types.Comet{}, errors.NewStoreErr(errors.KeyNotFound, key)
	}

	return res.(types.Comet), nil
}

func (s *InmemStore) SetComet(event types.Comet) error {
	key := event.Hex()
	_, err := s.GetComet(key)
	if err != nil && !errors.Is(err, errors.KeyNotFound) {
		return err
	}
	if errors.Is(err, errors.KeyNotFound) {
		if err := s.addParticpantEvent(event.Creator(), key, event.Index()); err != nil {
			return err
		}
	}
	s.eventCache.Add(key, event)

	return nil
}

func (s *InmemStore) addParticpantEvent(participant string, hash string, index int) error {
	return s.participantEventsCache.Set(participant, hash, index)
}

func (s *InmemStore) ParticipantEvents(participant string, skip int) ([]string, error) {
	return s.participantEventsCache.Get(participant, skip)
}

func (s *InmemStore) ParticipantEvent(particant string, index int) (string, error) {
	return s.participantEventsCache.GetItem(particant, index)
}

func (s *InmemStore) LastEventFrom(participant string) (last string, isRoot bool, err error) {
	//try to get the last event from this participant
	last, err = s.participantEventsCache.GetLast(participant)
	if err != nil {
		return
	}
	//if there is none, grab the root
	if last == "" {
		root, ok := s.roots[participant]
		if ok {
			last = root.X
			isRoot = true
		} else {
			err = errors.NewStoreErr(errors.NoRoot, participant)
		}
	}
	return
}

func (s *InmemStore) KnownEvents() map[int]int {
	return s.participantEventsCache.Known()
}

func (s *InmemStore) ConsensusEvents() []string {
	lastWindow, _ := s.consensusCache.GetLastWindow()
	res := make([]string, len(lastWindow))
	for i, item := range lastWindow {
		res[i] = item.(string)
	}
	return res
}

func (s *InmemStore) ConsensusEventsCount() int {
	return s.totConsensusEvents
}

func (s *InmemStore) AddConsensusEvent(key string) error {
	s.consensusCache.Set(key, s.totConsensusEvents)
	s.totConsensusEvents++
	return nil
}

func (s *InmemStore) GetRound(r int) (types.RoundInfo, error) {
	res, ok := s.roundCache.Get(r)
	if !ok {
		return *types.NewRoundInfo(), errors.NewStoreErr(errors.KeyNotFound, strconv.Itoa(r))
	}
	return res.(types.RoundInfo), nil
}

func (s *InmemStore) SetRound(r int, round types.RoundInfo) error {
	s.roundCache.Add(r, round)
	if r > s.lastRound {
		s.lastRound = r
	}
	return nil
}

func (s *InmemStore) LastRound() int {
	return s.lastRound
}

func (s *InmemStore) RoundWitnesses(r int) []string {
	round, err := s.GetRound(r)
	if err != nil {
		return []string{}
	}
	return round.Witnesses()
}

func (s *InmemStore) RoundEvents(r int) int {
	round, err := s.GetRound(r)
	if err != nil {
		return 0
	}
	return len(round.Events)
}

func (s *InmemStore) GetRoot(participant string) (types.Root, error) {
	res, ok := s.roots[participant]
	if !ok {
		return types.Root{}, errors.NewStoreErr(errors.KeyNotFound, participant)
	}
	return res, nil
}

func (s *InmemStore) GetBlock(index int) (types.Block, error) {
	res, ok := s.blockCache.Get(index)
	if !ok {
		return types.Block{}, errors.NewStoreErr(errors.KeyNotFound, strconv.Itoa(index))
	}
	return res.(types.Block), nil
}

func (s *InmemStore) SetBlock(block types.Block) error {
	_, err := s.GetBlock(block.Index())
	if err != nil && !errors.Is(err, errors.KeyNotFound) {
		return err
	}
	s.blockCache.Add(block.Index(), block)
	return nil
}

func (s *InmemStore) Reset(roots map[string]types.Root) error {
	s.roots = roots
	s.eventCache = common.NewLRU(s.cacheSize, nil)
	s.roundCache = common.NewLRU(s.cacheSize, nil)
	s.consensusCache = common.NewRollingIndex(s.cacheSize)
	err := s.participantEventsCache.Reset()
	s.lastRound = -1
	return err
}

func (s *InmemStore) Close() error {
	return nil
}
