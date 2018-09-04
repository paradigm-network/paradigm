package storage

import "github.com/paradigm-network/paradigm/types"

type Store interface {
	CacheSize() int
	Participants() (map[string]int, error)
	GetComet(string) (types.Comet, error)
	SetComet(types.Comet) error
	ParticipantEvents(string, int) ([]string, error)
	ParticipantEvent(string, int) (string, error)
	LastEventFrom(string) (string, bool, error)
	KnownEvents() map[int]int
	ConsensusEvents() []string
	ConsensusEventsCount() int
	AddConsensusEvent(string) error
	GetRound(int) (types.RoundInfo, error)
	SetRound(int, types.RoundInfo) error
	LastRound() int
	RoundWitnesses(int) []string
	RoundEvents(int) int
	GetRoot(string) (types.Root, error)
	GetBlock(int) (types.Block, error)
	SetBlock(types.Block) error
	Reset(map[string]types.Root) error
	Close() error
}

