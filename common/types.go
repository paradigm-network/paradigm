package common

const (
	AddressLength = 20
	HashLength    = 32
)

// Address represents the 20 byte address of an Paradigm account.
type Address [AddressLength]byte

// Hash represents the 32 byte Keccak256 hash of arbitrary data.
type Hash [HashLength]byte

// BytesToAddress returns Address with value b.
// If b is larger than len(h), b will be cropped from the left.
func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

// SetBytes sets the address to the value of b.
// If b is larger than len(a) it will panic.
func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
}

func HexToAddress(s string) Address { return BytesToAddress(FromHex(s)) }
