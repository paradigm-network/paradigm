package proxy

import "errors"

var (
	ErrNonceTooHigh              = errors.New("nonce is too high")
	ErrNonceTooLow               = errors.New("nonce is too low")
	ErrInsufficientBalanceForGas = errors.New("insufficient balance to pay for gas")
	ErrInsufficientBalance       = errors.New("insufficient balance")
	ErrOutOfGas                  = errors.New("out of gas")
	ErrGasLimitReached           = errors.New("gas limit reached")
)
