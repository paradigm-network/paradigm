package sequentia

import (
	"github.com/paradigm-network/paradigm/storage"
	"github.com/paradigm-network/paradigm/common"
	"github.com/paradigm-network/paradigm/types"
)

type CometGraph struct {
	Participants            map[string]int //[public key] => id
	ReverseParticipants     map[int]string //[id] => public key
	Store                   storage.Store          //store of Events and Rounds
	UndeterminedEvents      []string       //[index] => hash
	UndecidedRounds         []int          //queue of Rounds which have undecided witnesses
	LastConsensusRound      *int           //index of last round where the fame of all witnesses has been decided
	LastBlockIndex          int            //index of last block
	LastCommitedRoundEvents int            //number of events in round before LastConsensusRound
	ConsensusTransactions   int            //number of consensus transactions
	PendingLoadedEvents     int            //number of loaded events that are not yet committed
	commitCh                chan types.Block     //channel for committing Blocks
	topologicalIndex        int            //counter used to order events in topological order
	mostPlurality           int

	ancestorCache           *common.LRU
	selfAncestorCache       *common.LRU
	oldestSelfAncestorCache *common.LRU
	stronglySeeCache        *common.LRU
	parentRoundCache        *common.LRU
	roundCache              *common.LRU
}

func NewCometGraph(participants map[string]int, store storage.Store, commitCh chan types.Block) *CometGraph {
	reverseParticipants := make(map[int]string)
	for pk, id := range participants {
		reverseParticipants[id] = pk
	}

	cacheSize := store.CacheSize()
	cometGraph := CometGraph{
		Participants:            participants,
		ReverseParticipants:     reverseParticipants,
		Store:                   store,
		commitCh:                commitCh,
		ancestorCache:           common.NewLRU(cacheSize, nil),
		selfAncestorCache:       common.NewLRU(cacheSize, nil),
		oldestSelfAncestorCache: common.NewLRU(cacheSize, nil),
		stronglySeeCache:        common.NewLRU(cacheSize, nil),
		parentRoundCache:        common.NewLRU(cacheSize, nil),
		roundCache:              common.NewLRU(cacheSize, nil),
		mostPlurality:           2*len(participants)/3 + 1,
		UndecidedRounds:         []int{0}, //initialize,
		LastBlockIndex:          -1,
	}

	return &cometGraph
}

func (cg *CometGraph) MostPlurality() int {
	return cg.mostPlurality
}

//true if y is an ancestor of x
func (cg *CometGraph) AncestorOf(x, y string) bool {
	if c, ok := cg.ancestorCache.Get(storage.Key{x, y}); ok {
		return c.(bool)
	}
	a := cg.ancestor(x, y)
	cg.ancestorCache.Add(storage.Key{x, y}, a)
	return a
}

func (cg *CometGraph) ancestor(x, y string) bool {
	if x == y {
		return true
	}

	ex, err := cg.Store.GetComet(x)
	if err != nil {
		return false
	}

	ey, err := cg.Store.GetComet(y)
	if err != nil {
		return false
	}

	eyCreator := cg.Participants[ey.Creator()]
	lastAncestorKnownFromYCreator := ex.LastAncestors[eyCreator].Index

	return lastAncestorKnownFromYCreator >= ey.Index()
}

//true if y is a self-ancestor of x
func (cg *CometGraph) SelfAncestor(x, y string) bool {
	if c, ok := cg.selfAncestorCache.Get(storage.Key{x, y}); ok {
		return c.(bool)
	}
	a := cg.selfAncestor(x, y)
	cg.selfAncestorCache.Add(storage.Key{x, y}, a)
	return a
}

func (cg *CometGraph) selfAncestor(x, y string) bool {
	if x == y {
		return true
	}
	ex, err := cg.Store.GetComet(x)
	if err != nil {
		return false
	}
	exCreator := cg.Participants[ex.Creator()]

	ey, err := cg.Store.GetComet(y)
	if err != nil {
		return false
	}
	eyCreator := cg.Participants[ey.Creator()]

	return exCreator == eyCreator && ex.Index() >= ey.Index()
}

