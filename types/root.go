package types

import (
	"bytes"
	"encoding/json"
)

type Root struct {
	X, Y   string
	Index  int
	Round  int
	Others map[string]string
}

func NewBaseRoot() Root {
	return Root{
		X:     "",
		Y:     "",
		Index: -1,
		Round: -1,
	}
}

func (root *Root) Marshal() ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b) //will write to b
	if err := enc.Encode(root); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (root *Root) Unmarshal(data []byte) error {
	b := bytes.NewBuffer(data)
	dec := json.NewDecoder(b) //will read from b
	return dec.Decode(root)
}
