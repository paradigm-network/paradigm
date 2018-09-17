package types

import (
	"crypto/ecdsa"
	"math/big"
)


// Signer encapsulates transaction signature handling. Note that this interface is not a
// stable API and may change at any time to accommodate new protocol rules.
type Signer interface {

}

type EIP155Signer struct {

}

type HomesteadSigner struct{  }




// SignTx signs the transaction using the given signer and private key
func SignTx(tx *Transaction, s Signer, prv *ecdsa.PrivateKey) (*Transaction, error) {

	return nil, nil
}


func NewEIP155Signer(chainId *big.Int) EIP155Signer {
	return EIP155Signer{}
}


