package accounts

import (
	"errors"
	"fmt"
)



// ErrUnknownWallet is returned for any requested operation for which no backend
// provides the specified wallet.
var ErrUnknownWallet = errors.New("unknown wallet")


// ErrUnknownAccount is returned for any requested operation for which no backend
// provides the specified account.
var ErrUnknownAccount = errors.New("unknown account")


// ErrNotSupported is returned when an operation is requested from an account
// backend that it does not support.
var ErrNotSupported = errors.New("not supported")


// NewAuthNeededError creates a new authentication error with the extra details
// about the needed fields set.
func NewAuthNeededError(needed string) error {
	return &AuthNeededError{
		Needed: needed,
	}
}

// AuthNeededError is returned by backends for signing requests where the user
// is required to provide further authentication before signing can succeed.
//
// This usually means either that a password needs to be supplied, or perhaps a
// one time PIN code displayed by some hardware device.
type AuthNeededError struct {
	Needed string // Extra authentication the user needs to provide
}

// Error implements the standard error interface.
func (err *AuthNeededError) Error() string {
	return fmt.Sprintf("authentication needed: %s", err.Needed)
}
