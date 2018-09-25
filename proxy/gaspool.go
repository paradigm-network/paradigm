package proxy

import "math/big"

// GasPool tracks the amount of gas available during
// execution of the transactions in a block.
// The zero value is a pool with zero gas available.
type GasPool big.Int

// AddGas makes gas available for execution.
func (gp *GasPool) AddGas(amount *big.Int) *GasPool {
	i := (*big.Int)(gp)
	i.Add(i, amount)
	return gp
}

// SubGas deducts the given amount from the pool if enough gas is
// available and returns an error otherwise.
func (gp *GasPool) SubGas(amount *big.Int) error {
	i := (*big.Int)(gp)
	if i.Cmp(amount) < 0 {
		return ErrGasLimitReached
	}
	i.Sub(i, amount)
	return nil
}

func (gp *GasPool) String() string {
	return (*big.Int)(gp).String()
}
