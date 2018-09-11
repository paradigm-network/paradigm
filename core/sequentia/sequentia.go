package sequentia

import (
	"fmt"
	"math"
	"encoding/hex"
	"sort"
	"time"
	"github.com/paradigm-network/paradigm/storage"
	"github.com/paradigm-network/paradigm/common"
	"github.com/paradigm-network/paradigm/types"
	"github.com/paradigm-network/paradigm/errors"
	"sync/atomic"
)

// CometGraph is core component of Sequentia layer.
type CometGraph struct {
	Participants            map[string]int   //map of all node running the sequentia layer  [public key] => id
	ReverseParticipants     map[int]string   //reverse of Participants map  [id] => public key
	Store                   storage.Store    //storage interface of Comets and Comets Rounds
	UndeterminedEvents      []string         //undetermined comets [index] => hash
	UndecidedRounds         []int            //queue of Rounds which have undecided witnesses
	LastConsensusRound      *int             //index of last round where the fame of all witnesses has been decided
	LastBlockIndex          int              //index of last block
	LastCommitedRoundEvents int              //number of events in round before LastConsensusRound
	ConsensusTransactions   int              //number of consensus transactions
	PendingLoadedEvents     int              //number of loaded events that are not yet committed
	topologicalIndex        int              //counter used to order events in topological order
	mostPlurality           int

	commitCh                chan types.Block //channel for committing Blocks

	//caches
	ancestorCache           *common.LRU
	selfAncestorCache       *common.LRU
	oldestSelfAncestorCache *common.LRU
	stronglySeeCache        *common.LRU
	parentRoundCache        *common.LRU
	roundCache              *common.LRU
}

// Global Instance of CometGraph.
var Instance atomic.Value

// Build a new CometGraph struct.
func BuildCometGraph(participants map[string]int, store storage.Store, commitCh chan types.Block) *CometGraph {
	if Instance.Load() != nil {
		return Instance.Load().(*CometGraph)
	}
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
	Instance.Store(cometGraph)
	return &cometGraph
}

func (cg *CometGraph) MostPlurality() int {
	return cg.mostPlurality
}

//true if y is an ancestor of x
func (cg *CometGraph) AncestorOf(cx, cy string) bool {
	if c, ok := cg.ancestorCache.Get(storage.NewKey(cx, cy)); ok {
		return c.(bool)
	}
	a := cg.ancestor(cx, cy)
	cg.ancestorCache.Add(storage.NewKey(cx, cy), a)
	return a
}

//true if y is an ancestor of x
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
	if c, ok := cg.selfAncestorCache.Get(storage.NewKey(x, y)); ok {
		return c.(bool)
	}
	a := cg.selfAncestor(x, y)
	cg.selfAncestorCache.Add(storage.NewKey(x, y), a)
	return a
}

//true if y is a self-ancestor of x
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

//true if x sees y
func (cg *CometGraph) See(x, y string) bool {
	return cg.AncestorOf(x, y)
	//it is not necessary to detect forks because we assume that with our
	//implementations, no two events can be added by the same creator at the
	//same height (cf InsertComet)
}

//oldest self-ancestor of x to see y
func (cg *CometGraph) OldestSelfAncestorToSee(x, y string) string {
	if c, ok := cg.oldestSelfAncestorCache.Get(storage.NewKey(x, y)); ok {
		return c.(string)
	}
	res := cg.oldestSelfAncestorToSee(x, y)
	cg.oldestSelfAncestorCache.Add(storage.NewKey(x, y), res)
	return res
}

func (cg *CometGraph) oldestSelfAncestorToSee(x, y string) string {
	ex, err := cg.Store.GetComet(x)
	if err != nil {
		return ""
	}
	ey, err := cg.Store.GetComet(y)
	if err != nil {
		return ""
	}

	a := ey.FirstDescendants[cg.Participants[ex.Creator()]]

	if a.Index <= ex.Index() {
		return a.Hash
	}

	return ""
}

//true if x strongly sees y
func (cg *CometGraph) StronglySee(x, y string) bool {
	if c, ok := cg.stronglySeeCache.Get(storage.NewKey(x, y)); ok {
		return c.(bool)
	}
	ss := cg.stronglySee(x, y)
	cg.stronglySeeCache.Add(storage.NewKey(x, y), ss)
	return ss
}

func (cg *CometGraph) stronglySee(x, y string) bool {

	ex, err := cg.Store.GetComet(x)
	if err != nil {
		return false
	}

	ey, err := cg.Store.GetComet(y)
	if err != nil {
		return false
	}

	c := 0
	for i := 0; i < len(ex.LastAncestors); i++ {
		if ex.LastAncestors[i].Index >= ey.FirstDescendants[i].Index {
			c++
		}
	}
	return c >= cg.MostPlurality()
}

//PRI.round: max of parent rounds
//PRI.isRoot: true if round is taken from a Root
func (cg *CometGraph) ParentRound(x string) storage.ParentRoundInfo {
	if c, ok := cg.parentRoundCache.Get(x); ok {
		return c.(storage.ParentRoundInfo)
	}
	pr := cg.parentRound(x)
	cg.parentRoundCache.Add(x, pr)
	return pr
}

func (cg *CometGraph) parentRound(x string) storage.ParentRoundInfo {
	res := storage.NewBaseParentRoundInfo()

	ex, err := cg.Store.GetComet(x)
	if err != nil {
		return res
	}

	//We are going to need the Root later
	root, err := cg.Store.GetRoot(ex.Creator())
	if err != nil {
		return res
	}

	spRound := -1
	spRoot := false
	//If it is the creator's first Event, use the corresponding Root
	if ex.SelfParent() == root.X {
		spRound = root.Round
		spRoot = true
	} else {
		spRound = cg.Round(ex.SelfParent())
		spRoot = false
	}

	opRound := -1
	opRoot := false
	if _, err := cg.Store.GetComet(ex.OtherParent()); err == nil {
		//if we known the other-parent, fetch its Round directly
		opRound = cg.Round(ex.OtherParent())
	} else if ex.OtherParent() == root.Y {
		//we do not know the other-parent but it is referenced in Root.Y
		opRound = root.Round
		opRoot = true
	} else if other, ok := root.Others[x]; ok && other == ex.OtherParent() {
		//we do not know the other-parent but it is referenced  in Root.Others
		//we use the Root's Round
		//in reality the OtherParent Round is not necessarily the same as the
		//Root's but it is necessarily smaller. Since We are intererest in the
		//max between self-parent and other-parent rounds, this shortcut is
		//acceptable.
		opRound = root.Round
	}

	res.Round = spRound
	res.IsRoot = spRoot
	if spRound < opRound {
		res.Round = opRound
		res.IsRoot = opRoot
	}
	return res
}

//true if x is a witness (first event of a round for the owner)
func (cg *CometGraph) Witness(x string) bool {
	ex, err := cg.Store.GetComet(x)
	if err != nil {
		return false
	}

	root, err := cg.Store.GetRoot(ex.Creator())
	if err != nil {
		return false
	}

	//If it is the creator's first Event, return true
	if ex.SelfParent() == root.X && ex.OtherParent() == root.Y {
		return true
	}

	return cg.Round(x) > cg.Round(ex.SelfParent())
}

//true if round of x should be incremented
func (cg *CometGraph) RoundInc(x string) bool {

	parentRound := cg.ParentRound(x)

	//If parent-round was obtained from a Root, then x is the Event that sits
	//right on top of the Root. RoundInc is true.
	if parentRound.IsRoot {
		return true
	}

	//If parent-round was obtained from a regulare Event, then we need to check
	//if x strongly-sees a strong majority of withnesses from parent-round.
	c := 0
	for _, w := range cg.Store.RoundWitnesses(parentRound.Round) {
		if cg.StronglySee(x, w) {
			c++
		}
	}

	return c >= cg.MostPlurality()
}

func (cg *CometGraph) RoundReceived(x string) int {

	ex, err := cg.Store.GetComet(x)
	if err != nil {
		return -1
	}
	if ex.RoundReceived == nil {
		return -1
	}

	return *ex.RoundReceived
}

func (cg *CometGraph) Round(x string) int {
	if c, ok := cg.roundCache.Get(x); ok {
		return c.(int)
	}
	r := cg.round(x)
	cg.roundCache.Add(x, r)
	return r
}

func (cg *CometGraph) round(x string) int {

	round := cg.ParentRound(x).Round

	inc := cg.RoundInc(x)

	if inc {
		round++
	}
	return round
}

//round(x) - round(y)
func (cg *CometGraph) RoundDiff(x, y string) (int, error) {

	xRound := cg.Round(x)
	if xRound < 0 {
		return math.MinInt32, fmt.Errorf("event %s has negative round", x)
	}
	yRound := cg.Round(y)
	if yRound < 0 {
		return math.MinInt32, fmt.Errorf("event %s has negative round", y)
	}

	return xRound - yRound, nil
}

//insert comet into db, with comet check and wireInfo if setWireInfo is true.
func (cg *CometGraph) InsertComet(comet types.Comet, setWireInfo bool) error {
	//verify signature
	if ok, err := comet.Verify(); !ok {
		if err != nil {
			return err
		}
		return fmt.Errorf("Invalid Event signature")
	}

	if err := cg.CheckSelfParent(comet); err != nil {
		return fmt.Errorf("CheckSelfParent: %s", err)
	}

	if err := cg.CheckOtherParent(comet); err != nil {
		return fmt.Errorf("CheckOtherParent: %s", err)
	}

	comet.TopologicalIndex = cg.topologicalIndex
	cg.topologicalIndex++

	if setWireInfo {
		if err := cg.SetWireInfo(&comet); err != nil {
			return fmt.Errorf("SetWireInfo: %s", err)
		}
	}

	if err := cg.InitEventCoordinates(&comet); err != nil {
		return fmt.Errorf("InitEventCoordinates: %s", err)
	}

	if err := cg.Store.SetComet(comet); err != nil {
		return fmt.Errorf("SetEvent: %s", err)
	}

	if err := cg.UpdateAncestorFirstDescendant(comet); err != nil {
		return fmt.Errorf("UpdateAncestorFirstDescendant: %s", err)
	}

	cg.UndeterminedEvents = append(cg.UndeterminedEvents, comet.Hex())

	if comet.IsLoaded() {
		cg.PendingLoadedEvents++
	}

	cg.recordBlockSignatures(comet.BlockSignatures())

	return nil
}

func (cg *CometGraph) recordBlockSignatures(blockSignatures []types.BlockSignature) {
	for _, bs := range blockSignatures {
		//check if validator belongs to list of participants
		validatorHex := fmt.Sprintf("0x%X", bs.Validator)
		if _, ok := cg.Participants[validatorHex]; !ok {
			//cg.logger.WithFields(logrus.Fields{
			//	"index":     bs.Index,
			//	"validator": validatorHex,
			//}).Warning("Verifying Block signature. Unknown validator")
			continue
		}

		block, err := cg.Store.GetBlock(bs.Index)
		if err != nil {
			//h.logger.WithFields(logrus.Fields{
			//	"index": bs.Index,
			//	"msg":   err,
			//}).Warning("Verifying Block signature. Could not fetch Block")
			continue
		}
		valid, err := block.Verify(bs)
		if err != nil {
			//h.logger.WithFields(logrus.Fields{
			//	"index": bs.Index,
			//	"msg":   err,
			//}).Warning("Verifying Block signature")
			continue
		}
		if !valid {
			//h.logger.WithFields(logrus.Fields{
			//	"index": bs.Index,
			//}).Warning("Verifying Block signature. Invalid signature")
			continue
		}

		block.SetSignature(bs)

		if err := cg.Store.SetBlock(block); err != nil {
			//h.logger.WithFields(logrus.Fields{
			//	"index": bs.Index,
			//	"msg":   err,
			//}).Warning("Saving Block")
		}
	}
}

//Check the SelfParent is the Creator's last known comet.
func (cg *CometGraph) CheckSelfParent(comet types.Comet) error {
	selfParent := comet.SelfParent()
	creator := comet.Creator()

	creatorLastKnown, _, err := cg.Store.LastEventFrom(creator)
	if err != nil {
		return err
	}

	selfParentLegit := selfParent == creatorLastKnown

	if !selfParentLegit {
		return fmt.Errorf("Self-parent not last known event by creator")
	}

	return nil
}

//Check if we know the OtherParent
func (cg *CometGraph) CheckOtherParent(comet types.Comet) error {
	otherParent := comet.OtherParent()
	if otherParent != "" {
		//Check if we have it
		_, err := cg.Store.GetComet(otherParent)
		if err != nil {
			//it might still be in the Root
			root, err := cg.Store.GetRoot(comet.Creator())
			if err != nil {
				return err
			}
			if root.X == comet.SelfParent() && root.Y == otherParent {
				return nil
			}
			other, ok := root.Others[comet.Hex()]
			if ok && other == comet.OtherParent() {
				return nil
			}
			return fmt.Errorf("Other-parent not known")
		}
	}
	return nil
}

//initialize arrays of last ancestors and first descendants
func (cg *CometGraph) InitEventCoordinates(comet *types.Comet) error {
	members := len(cg.Participants)

	comet.FirstDescendants = make([]types.EventCoordinates, members)
	for id := 0; id < members; id++ {
		comet.FirstDescendants[id] = types.EventCoordinates{
			Index: math.MaxInt32,
		}
	}

	comet.LastAncestors = make([]types.EventCoordinates, members)

	selfParent, selfParentError := cg.Store.GetComet(comet.SelfParent())
	otherParent, otherParentError := cg.Store.GetComet(comet.OtherParent())

	if selfParentError != nil && otherParentError != nil {
		for id := 0; id < members; id++ {
			comet.LastAncestors[id] = types.EventCoordinates{
				Index: -1,
			}
		}
	} else if selfParentError != nil {
		copy(comet.LastAncestors[:members], otherParent.LastAncestors)
	} else if otherParentError != nil {
		copy(comet.LastAncestors[:members], selfParent.LastAncestors)
	} else {
		selfParentLastAncestors := selfParent.LastAncestors
		otherParentLastAncestors := otherParent.LastAncestors

		copy(comet.LastAncestors[:members], selfParentLastAncestors)
		for i := 0; i < members; i++ {
			if comet.LastAncestors[i].Index < otherParentLastAncestors[i].Index {
				comet.LastAncestors[i].Index = otherParentLastAncestors[i].Index
				comet.LastAncestors[i].Hash = otherParentLastAncestors[i].Hash
			}
		}
	}

	index := comet.Index()

	creator := comet.Creator()
	creatorID, ok := cg.Participants[creator]
	if !ok {
		return fmt.Errorf("Could not find creator id (%s)", creator)
	}
	hash := comet.Hex()

	comet.FirstDescendants[creatorID] = types.EventCoordinates{Index: index, Hash: hash}
	comet.LastAncestors[creatorID] = types.EventCoordinates{Index: index, Hash: hash}

	return nil
}

//update first decendant of each last ancestor to point to comet
func (cg *CometGraph) UpdateAncestorFirstDescendant(comet types.Comet) error {
	creatorID, ok := cg.Participants[comet.Creator()]
	if !ok {
		return fmt.Errorf("Could not find creator id (%s)", comet.Creator())
	}
	index := comet.Index()
	hash := comet.Hex()

	for i := 0; i < len(comet.LastAncestors); i++ {
		ah := comet.LastAncestors[i].Hash
		for ah != "" {
			a, err := cg.Store.GetComet(ah)
			if err != nil {
				break
			}
			if a.FirstDescendants[creatorID].Index == math.MaxInt32 {
				a.FirstDescendants[creatorID] = types.EventCoordinates{Index: index, Hash: hash}
				if err := cg.Store.SetComet(a); err != nil {
					return err
				}
				ah = a.SelfParent()
			} else {
				break
			}
		}
	}

	return nil
}

func (cg *CometGraph) SetWireInfo(comet *types.Comet) error {
	selfParentIndex := -1
	otherParentCreatorID := -1
	otherParentIndex := -1

	//could be the first Event inserted for this creator. In this case, use Root
	if lf, isRoot, _ := cg.Store.LastEventFrom(comet.Creator()); isRoot && lf == comet.SelfParent() {
		root, err := cg.Store.GetRoot(comet.Creator())
		if err != nil {
			return err
		}
		selfParentIndex = root.Index
	} else {
		selfParent, err := cg.Store.GetComet(comet.SelfParent())
		if err != nil {
			return err
		}
		selfParentIndex = selfParent.Index()
	}

	if comet.OtherParent() != "" {
		otherParent, err := cg.Store.GetComet(comet.OtherParent())
		if err != nil {
			return err
		}
		otherParentCreatorID = cg.Participants[otherParent.Creator()]
		otherParentIndex = otherParent.Index()
	}

	comet.SetWireInfo(selfParentIndex,
		otherParentCreatorID,
		otherParentIndex,
		cg.Participants[comet.Creator()])

	return nil
}

func (cg *CometGraph) ReadWireInfo(wevent types.WireEvent) (*types.Comet, error) {
	selfParent := ""
	otherParent := ""
	var err error

	creator := cg.ReverseParticipants[wevent.Body.CreatorID]
	creatorBytes, err := hex.DecodeString(creator[2:])
	if err != nil {
		return nil, err
	}

	if wevent.Body.SelfParentIndex >= 0 {
		selfParent, err = cg.Store.ParticipantEvent(creator, wevent.Body.SelfParentIndex)
		if err != nil {
			return nil, err
		}
	}
	if wevent.Body.OtherParentIndex >= 0 {
		otherParentCreator := cg.ReverseParticipants[wevent.Body.OtherParentCreatorID]
		otherParent, err = cg.Store.ParticipantEvent(otherParentCreator, wevent.Body.OtherParentIndex)
		if err != nil {
			return nil, err
		}
	}

	body := types.CometBody{
		Transactions:    wevent.Body.Transactions,
		BlockSignatures: wevent.BlockSignatures(creatorBytes),
		Parents:         []string{selfParent, otherParent},
		Creator:         creatorBytes,

		Timestamp:            wevent.Body.Timestamp,
		Index:                wevent.Body.Index,
		SelfParentIndex:      wevent.Body.SelfParentIndex,
		OtherParentCreatorID: wevent.Body.OtherParentCreatorID,
		OtherParentIndex:     wevent.Body.OtherParentIndex,
		CreatorID:            wevent.Body.CreatorID,
	}

	comet := &types.Comet{
		Body:      body,
		Signature: wevent.Signature,
	}

	return comet, nil
}

func (cg *CometGraph) DivideRounds() error {
	for _, hash := range cg.UndeterminedEvents {
		roundNumber := cg.Round(hash)
		witness := cg.Witness(hash)
		roundInfo, err := cg.Store.GetRound(roundNumber)

		//If the RoundInfo is not found in the Store's Cache, then the Sequentia
		//is not aware of it yet. We need to add the roundNumber to the queue of
		//undecided rounds so that it will be processed in the other consensus
		//methods
		if err != nil && !errors.Is(err, errors.KeyNotFound) {
			return err
		}
		//If the RoundInfo is actually taken from the Store's DB, then it still
		//has not been processed by the Sequentia consensus methods (The 'queued'
		//field is not exported and therefore not persisted in the DB).
		//RoundInfos taken from the DB directly will always have this field set
		//to false
		if !roundInfo.Queued {
			cg.UndecidedRounds = append(cg.UndecidedRounds, roundNumber)
			roundInfo.Queued = true
		}

		roundInfo.AddEvent(hash, witness)
		err = cg.Store.SetRound(roundNumber, roundInfo)
		if err != nil {
			return err
		}
	}
	return nil
}

//decide if witnesses are famous
func (cg *CometGraph) DecideFame() error {
	votes := make(map[string]map[string]bool) //[x][y]=>vote(x,y)

	decidedRounds := map[int]int{} // [round number] => index in h.UndecidedRounds
	defer cg.updateUndecidedRounds(decidedRounds)

	for pos, i := range cg.UndecidedRounds {
		roundInfo, err := cg.Store.GetRound(i)
		if err != nil {
			return err
		}
		for _, x := range roundInfo.Witnesses() {
			if roundInfo.IsDecided(x) {
				continue
			}
		X:
			for j := i + 1; j <= cg.Store.LastRound(); j++ {
				for _, y := range cg.Store.RoundWitnesses(j) {
					diff := j - i
					if diff == 1 {
						setVote(votes, y, x, cg.See(y, x))
					} else {
						//count votes
						ssWitnesses := []string{}
						for _, w := range cg.Store.RoundWitnesses(j - 1) {
							if cg.StronglySee(y, w) {
								ssWitnesses = append(ssWitnesses, w)
							}
						}
						yays := 0
						nays := 0
						for _, w := range ssWitnesses {
							if votes[w][x] {
								yays++
							} else {
								nays++
							}
						}
						v := false
						t := nays
						if yays >= nays {
							v = true
							t = yays
						}

						//normal round
						if math.Mod(float64(diff), float64(len(cg.Participants))) > 0 {
							if t >= cg.MostPlurality() {
								roundInfo.SetFame(x, v)
								setVote(votes, y, x, v)
								break X //break out of j loop
							} else {
								setVote(votes, y, x, v)
							}
						} else { //coin round
							if t >= cg.MostPlurality() {
								setVote(votes, y, x, v)
							} else {
								setVote(votes, y, x, middleBit(y)) //middle bit of y's hash
							}
						}
					}
				}
			}
		}

		//Update decidedRounds and LastConsensusRound if all witnesses have been decided
		if roundInfo.WitnessesDecided() {
			decidedRounds[i] = pos

			if cg.LastConsensusRound == nil || i > *cg.LastConsensusRound {
				cg.setLastConsensusRound(i)
			}
		}

		err = cg.Store.SetRound(i, roundInfo)
		if err != nil {
			return err
		}
	}
	return nil
}

//remove items from UndecidedRounds
func (cg *CometGraph) updateUndecidedRounds(decidedRounds map[int]int) {
	var newUndecidedRounds []int
	for _, ur := range cg.UndecidedRounds {
		if _, ok := decidedRounds[ur]; !ok {
			newUndecidedRounds = append(newUndecidedRounds, ur)
		}
	}
	cg.UndecidedRounds = newUndecidedRounds
}

func (cg *CometGraph) setLastConsensusRound(i int) {
	if cg.LastConsensusRound == nil {
		cg.LastConsensusRound = new(int)
	}
	*cg.LastConsensusRound = i

	cg.LastCommitedRoundEvents = cg.Store.RoundEvents(i - 1)
}

//assign round received and timestamp to all events
func (cg *CometGraph) DecideRoundReceived() error {
	for _, x := range cg.UndeterminedEvents {
		r := cg.Round(x)
		for i := r + 1; i <= cg.Store.LastRound(); i++ {
			tr, err := cg.Store.GetRound(i)
			if err != nil && !errors.Is(err, errors.KeyNotFound) {
				return err
			}

			//skip if some witnesses are left undecided
			if !(tr.WitnessesDecided() && cg.UndecidedRounds[0] > i) {
				continue
			}

			fws := tr.FamousWitnesses()
			//set of famous witnesses that see x
			var s []string
			for _, w := range fws {
				if cg.See(w, x) {
					s = append(s, w)
				}
			}
			if len(s) > len(fws)/2 {
				ex, err := cg.Store.GetComet(x)
				if err != nil {
					return err
				}
				ex.SetRoundReceived(i)

				t := []string{}
				for _, a := range s {
					t = append(t, cg.OldestSelfAncestorToSee(a, x))
				}

				ex.ConsensusTimestamp = cg.MedianTimestamp(t)

				err = cg.Store.SetComet(ex)
				if err != nil {
					return err
				}

				break
			}
		}
	}
	return nil
}

func (cg *CometGraph) FindOrder() error {
	err := cg.DecideRoundReceived()
	if err != nil {
		return err
	}

	var newConsensusEvents []types.Comet
	var newUndeterminedEvents []string
	for _, x := range cg.UndeterminedEvents {
		ex, err := cg.Store.GetComet(x)
		if err != nil {
			return err
		}
		if ex.RoundReceived != nil {
			newConsensusEvents = append(newConsensusEvents, ex)
		} else {
			newUndeterminedEvents = append(newUndeterminedEvents, x)
		}
	}
	cg.UndeterminedEvents = newUndeterminedEvents

	sorter := NewConsensusSorter(newConsensusEvents)
	sort.Sort(sorter)

	if err := cg.handleNewConsensusEvents(newConsensusEvents); err != nil {
		return err
	}

	return nil
}

func (cg *CometGraph) handleNewConsensusEvents(newConsensusEvents []types.Comet) error {

	blockMap := make(map[int][][]byte) // [RoundReceived] => []Transactions
	var blockOrder []int               // [index] => RoundReceived
	for _, e := range newConsensusEvents {
		err := cg.Store.AddConsensusEvent(e.Hex())
		if err != nil {
			return err
		}
		cg.ConsensusTransactions += len(e.Transactions())
		if e.IsLoaded() {
			cg.PendingLoadedEvents--
		}

		btxs, ok := blockMap[*e.RoundReceived]
		if !ok {
			btxs = [][]byte{}
			blockOrder = append(blockOrder, *e.RoundReceived)
		}
		btxs = append(btxs, e.Transactions()...)
		blockMap[*e.RoundReceived] = btxs
	}

	for _, rr := range blockOrder {
		blockTxs, _ := blockMap[rr]
		if len(blockTxs) > 0 {
			block, err := cg.createAndInsertBlock(rr, blockTxs)
			if err != nil {
				return err
			}
			if cg.commitCh != nil {
				cg.commitCh <- block
			}
		}
	}

	return nil
}

func (cg *CometGraph) createAndInsertBlock(roundReceived int, txs [][]byte) (types.Block, error) {
	block := types.NewBlock(cg.LastBlockIndex+1, roundReceived, txs)
	if err := cg.Store.SetBlock(block); err != nil {
		return types.Block{}, err
	}
	cg.LastBlockIndex++
	return block, nil
}

func (cg *CometGraph) MedianTimestamp(eventHashes []string) time.Time {
	var events []types.Comet
	for _, x := range eventHashes {
		ex, _ := cg.Store.GetComet(x)
		events = append(events, ex)
	}
	sort.Sort(types.ByTimestamp(events))
	return events[len(events)/2].Body.Timestamp
}

func (cg *CometGraph) ConsensusEvents() []string {
	return cg.Store.ConsensusEvents()
}

//last event index per participant
func (cg *CometGraph) KnownEvents() map[int]int {
	return cg.Store.KnownEvents()
}

func (cg *CometGraph) Reset(roots map[string]types.Root) error {
	if err := cg.Store.Reset(roots); err != nil {
		return err
	}

	cg.UndeterminedEvents = []string{}
	cg.UndecidedRounds = []int{}
	cg.PendingLoadedEvents = 0
	cg.topologicalIndex = 0

	cacheSize := cg.Store.CacheSize()
	cg.ancestorCache = common.NewLRU(cacheSize, nil)
	cg.selfAncestorCache = common.NewLRU(cacheSize, nil)
	cg.oldestSelfAncestorCache = common.NewLRU(cacheSize, nil)
	cg.stronglySeeCache = common.NewLRU(cacheSize, nil)
	cg.parentRoundCache = common.NewLRU(cacheSize, nil)
	cg.roundCache = common.NewLRU(cacheSize, nil)

	return nil
}

func (cg *CometGraph) GetFrame() (types.Frame, error) {
	lastConsensusRoundIndex := 0
	if lcr := cg.LastConsensusRound; lcr != nil {
		lastConsensusRoundIndex = *lcr
	}

	lastConsensusRound, err := cg.Store.GetRound(lastConsensusRoundIndex)
	if err != nil {
		return types.Frame{}, err
	}

	witnessHashes := lastConsensusRound.Witnesses()

	var events []types.Comet
	roots := make(map[string]types.Root)
	for _, wh := range witnessHashes {
		w, err := cg.Store.GetComet(wh)
		if err != nil {
			return types.Frame{}, err
		}
		events = append(events, w)
		roots[w.Creator()] = types.Root{
			X:      w.SelfParent(),
			Y:      w.OtherParent(),
			Index:  w.Index() - 1,
			Round:  cg.Round(w.SelfParent()),
			Others: map[string]string{},
		}

		participantEvents, err := cg.Store.ParticipantEvents(w.Creator(), w.Index())
		if err != nil {
			return types.Frame{}, err
		}
		for _, e := range participantEvents {
			ev, err := cg.Store.GetComet(e)
			if err != nil {
				return types.Frame{}, err
			}
			events = append(events, ev)
		}
	}

	//Not every participant necessarily has a witness in LastConsensusRound.
	//Hence, there could be participants with no Root at this point.
	//For these partcipants, use their last known Event.
	for p := range cg.Participants {
		if _, ok := roots[p]; !ok {
			var root types.Root
			last, isRoot, err := cg.Store.LastEventFrom(p)
			if err != nil {
				return types.Frame{}, err
			}
			if isRoot {
				root, err = cg.Store.GetRoot(p)
				if err != nil {
					return types.Frame{}, err
				}
			} else {
				ev, err := cg.Store.GetComet(last)
				if err != nil {
					return types.Frame{}, err
				}
				events = append(events, ev)
				root = types.Root{
					X:      ev.SelfParent(),
					Y:      ev.OtherParent(),
					Index:  ev.Index() - 1,
					Round:  cg.Round(ev.SelfParent()),
					Others: map[string]string{},
				}
			}
			roots[p] = root
		}
	}

	sort.Sort(types.ByTopologicalOrder(events))

	//Some Events in the Frame might have other-parents that are outside of the
	//Frame (cf root.go ex 2)
	//When inserting these Events in a newly reset Sequentia, the CheckOtherParent
	//method would return an error because the other-parent would not be found.
	//So we make it possible to also look for other-parents in the creator's Root.
	treated := map[string]bool{}
	for _, ev := range events {
		treated[ev.Hex()] = true
		otherParent := ev.OtherParent()
		if otherParent != "" {
			opt, ok := treated[otherParent]
			if !opt || !ok {
				if ev.SelfParent() != roots[ev.Creator()].X {
					roots[ev.Creator()].Others[ev.Hex()] = otherParent
				}
			}
		}
	}

	frame := types.Frame{
		Roots:  roots,
		Comets: events,
	}

	return frame, nil
}

//Bootstrap loads all Events from the Store's DB (if there is one) and feeds
//them to the Sequentia (in topological order) for consensus ordering. After this
//method call, the Sequentia should be in a state coeherent with the 'tip' of the
//Sequentia
func (cg *CometGraph) Bootstrap() error {
	if badgerStore, ok := cg.Store.(*storage.BadgerStore); ok {
		//Retreive the Events from the underlying DB. They come out in topological
		//order
		topologicalEvents, err := badgerStore.DbTopologicalEvents()
		if err != nil {
			return err
		}

		//Insert the Comets in the Sequentia
		for _, e := range topologicalEvents {
			if err := cg.InsertComet(e, true); err != nil {
				return err
			}
		}

		//Compute the consensus order of Events
		if err := cg.DivideRounds(); err != nil {
			return err
		}
		if err := cg.DecideFame(); err != nil {
			return err
		}
		if err := cg.FindOrder(); err != nil {
			return err
		}
	}

	return nil
}

func middleBit(ehex string) bool {
	hash, err := hex.DecodeString(ehex[2:])
	if err != nil {
		fmt.Printf("ERROR decoding hex string: %s\n", err)
	}
	if len(hash) > 0 && hash[len(hash)/2] == 0 {
		return false
	}
	return true
}

func setVote(votes map[string]map[string]bool, x, y string, vote bool) {
	if votes[x] == nil {
		votes[x] = make(map[string]bool)
	}
	votes[x][y] = vote
}
