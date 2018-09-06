package types

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"time"

	"github.com/paradigm-network/paradigm/common/crypto"
)

type CometBody struct {
	Transactions    [][]byte         //the payload
	Parents         []string         //hashes of the comet's parents, self-parent first
	Creator         []byte           //creator's public key
	Timestamp       time.Time        //creator's claimed timestamp of the comet's creation
	Index           int              //index in the sequence of comets created by Creator
	BlockSignatures []BlockSignature //list of Block signatures signed by the Comet's Creator ONLY

	//wire
	//It is cheaper to send ints then hashes over the wire
	selfParentIndex      int
	otherParentCreatorID int
	otherParentIndex     int
	creatorID            int
}

//json encoding of body only
func (e *CometBody) Marshal() ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b) //will write to b
	if err := enc.Encode(e); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *CometBody) Unmarshal(data []byte) error {
	b := bytes.NewBuffer(data)
	dec := json.NewDecoder(b) //will read from b
	if err := dec.Decode(e); err != nil {
		return err
	}
	return nil
}

func (e *CometBody) Hash() ([]byte, error) {
	hashBytes, err := e.Marshal()
	if err != nil {
		return nil, err
	}
	return crypto.SHA256(hashBytes), nil
}

type EventCoordinates struct {
	Hash  string
	Index int
}

type Comet struct {
	Body      CometBody
	Signature string //creator's digital signature of body

	TopologicalIndex int

	roundReceived      *int
	consensusTimestamp time.Time

	LastAncestors    []EventCoordinates //[participant fake id] => last ancestor
	FirstDescendants []EventCoordinates //[participant fake id] => first descendant

	creator string
	hash    []byte
	hex     string
}

func NewComet(transactions [][]byte,
	blockSignatures []BlockSignature,
	parents []string,
	creator []byte,
	index int) Comet {

	body := CometBody{
		Transactions:    transactions,
		BlockSignatures: blockSignatures,
		Parents:         parents,
		Creator:         creator,
		Timestamp:       time.Now().UTC(), //strip monotonic time
		Index:           index,
	}
	return Comet{
		Body: body,
	}
}

func (e *Comet) Creator() string {
	if e.creator == "" {
		e.creator = fmt.Sprintf("0x%X", e.Body.Creator)
	}
	return e.creator
}

func (e *Comet) SelfParent() string {
	return e.Body.Parents[0]
}

func (e *Comet) OtherParent() string {
	return e.Body.Parents[1]
}

func (e *Comet) Transactions() [][]byte {
	return e.Body.Transactions
}

func (e *Comet) Index() int {
	return e.Body.Index
}

func (e *Comet) BlockSignatures() []BlockSignature {
	return e.Body.BlockSignatures
}

//True if Comet contains a payload or is the initial Comet of its creator
func (e *Comet) IsLoaded() bool {
	if e.Body.Index == 0 {
		return true
	}

	hasTransactions := e.Body.Transactions != nil &&
		len(e.Body.Transactions) > 0

	hasBlockSignatures := e.Body.BlockSignatures != nil &&
		len(e.Body.BlockSignatures) > 0

	return hasTransactions || hasBlockSignatures
}

//ecdsa sig
func (e *Comet) Sign(privKey *ecdsa.PrivateKey) error {
	signBytes, err := e.Body.Hash()
	if err != nil {
		return err
	}
	R, S, err := crypto.Sign(privKey, signBytes)
	if err != nil {
		return err
	}
	e.Signature = crypto.EncodeSignature(R, S)
	return err
}

func (e *Comet) Verify() (bool, error) {
	pubBytes := e.Body.Creator
	pubKey := crypto.ToECDSAPub(pubBytes)

	signBytes, err := e.Body.Hash()
	if err != nil {
		return false, err
	}

	r, s, err := crypto.DecodeSignature(e.Signature)
	if err != nil {
		return false, err
	}

	return crypto.Verify(pubKey, signBytes, r, s), nil
}

//json encoding of body and signature
func (e *Comet) Marshal() ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	if err := enc.Encode(e); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (e *Comet) Unmarshal(data []byte) error {
	b := bytes.NewBuffer(data)
	dec := json.NewDecoder(b) //will read from b
	return dec.Decode(e)
}

//sha256 hash of body
func (e *Comet) Hash() ([]byte, error) {
	if len(e.hash) == 0 {
		hash, err := e.Body.Hash()
		if err != nil {
			return nil, err
		}
		e.hash = hash
	}
	return e.hash, nil
}

func (e *Comet) Hex() string {
	if e.hex == "" {
		hash, _ := e.Hash()
		e.hex = fmt.Sprintf("0x%X", hash)
	}
	return e.hex
}

func (e *Comet) SetRoundReceived(rr int) {
	if e.roundReceived == nil {
		e.roundReceived = new(int)
	}
	*e.roundReceived = rr
}

func (e *Comet) SetWireInfo(selfParentIndex,
	otherParentCreatorID,
	otherParentIndex,
	creatorID int) {
	e.Body.selfParentIndex = selfParentIndex
	e.Body.otherParentCreatorID = otherParentCreatorID
	e.Body.otherParentIndex = otherParentIndex
	e.Body.creatorID = creatorID
}

func (e *Comet) WireBlockSignatures() []WireBlockSignature {
	if e.Body.BlockSignatures != nil {
		wireSignatures := make([]WireBlockSignature, len(e.Body.BlockSignatures))
		for i, bs := range e.Body.BlockSignatures {
			wireSignatures[i] = bs.ToWire()
		}

		return wireSignatures
	}
	return nil
}

func (e *Comet) ToWire() WireEvent {

	return WireEvent{
		Body: WireBody{
			Transactions:         e.Body.Transactions,
			SelfParentIndex:      e.Body.selfParentIndex,
			OtherParentCreatorID: e.Body.otherParentCreatorID,
			OtherParentIndex:     e.Body.otherParentIndex,
			CreatorID:            e.Body.creatorID,
			Timestamp:            e.Body.Timestamp,
			Index:                e.Body.Index,
			BlockSignatures:      e.WireBlockSignatures(),
		},
		Signature: e.Signature,
	}
}

//Sorting

// ByTimestamp implements sort.Interface for []Comet based on
// the timestamp field.
type ByTimestamp []Comet

func (a ByTimestamp) Len() int      { return len(a) }
func (a ByTimestamp) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByTimestamp) Less(i, j int) bool {
	//normally, time.Sub uses monotonic time which only makes sense if it is
	//being called in the same process that made the time object.
	//that is why we strip out the monotonic time reading from the Timestamp at
	//the time of creating the Comet
	return a[i].Body.Timestamp.Before(a[j].Body.Timestamp)
}

// ByTopologicalOrder implements sort.Interface for []Comet based on
// the TopologicalIndex field.
type ByTopologicalOrder []Comet

func (a ByTopologicalOrder) Len() int      { return len(a) }
func (a ByTopologicalOrder) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByTopologicalOrder) Less(i, j int) bool {
	return a[i].TopologicalIndex < a[j].TopologicalIndex
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// WireEvent

type WireBody struct {
	Transactions    [][]byte
	BlockSignatures []WireBlockSignature

	SelfParentIndex      int
	OtherParentCreatorID int
	OtherParentIndex     int
	CreatorID            int

	Timestamp time.Time
	Index     int
}

type WireEvent struct {
	Body      WireBody
	Signature string
}

func (we *WireEvent) BlockSignatures(validator []byte) []BlockSignature {
	if we.Body.BlockSignatures != nil {
		blockSignatures := make([]BlockSignature, len(we.Body.BlockSignatures))
		for k, bs := range we.Body.BlockSignatures {
			blockSignatures[k] = BlockSignature{
				Validator: validator,
				Index:     bs.Index,
				Signature: bs.Signature,
			}
		}
		return blockSignatures
	}
	return nil
}
