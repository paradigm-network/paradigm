// Package accounts implements high level Paradigm account management.
package accounts

import (
	"math/big"
	"paradigm/common"
	"paradigm/core/types"
	"paradigm/event"
)

// Backend is a "wallet provider" that may contain a batch of accounts they can
// sign transactions with and upon request, do so.
type Backend interface {
	// Wallets retrieves the list of wallets the backend is currently aware of.
	Wallets() []Wallet

	// Subscribe creates an async subscription to receive notifications when the
	// backend detects the arrival or departure of a wallet.
	Subscribe(sink chan<- WalletEvent) event.Subscription
}

// Wallet represents a software or hardware wallet that might contain one or more
// accounts.
type Wallet interface {
	// URL retrieves the canonical path under which this wallet is reachable.
	URL() URL

	// Status returns a textual status. It also returns an error indicating any
	// failure the wallet might have encountered.
	Status() (string, error)

	// Open initializes access to a wallet instance. It is not meant to unlock or
	// decrypt account keys.
	Open(passphrase string) error

	// Close releases any resources held by an open wallet instance.
	Close() error

	// Accounts retrieves the list of signing accounts the wallet is currently aware of.
	Accounts() []Account

	// Contains returns whether an account is part of this particular wallet or not.
	Contains(account Account) bool

	// Derive attempts to explicitly derive a hierarchical deterministic account at
	// the specified derivation path.
	Derive(path DerivationPath, pin bool) (Account, error)

	// SignHash requests the wallet to sign the given hash.
	SignHash(account Account, hash []byte) ([]byte, error)

	// SignTx requests the wallet to sign the given transaction.
	SignTx(account Account, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error)

	// SignHashWithPassphrase requests the wallet to sign the given hash with the
	// given passphrase as extra authentication information.
	SignHashWithPassphrase(account Account, passphrase string, hash []byte) ([]byte, error)

	// SignTxWithPassphrase requests the wallet to sign the given transaction, with the
	// given passphrase as extra authentication information.
	SignTxWithPassphrase(account Account, passphrase string, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error)
}

// Account represents an Paradigm account located at a specific location defined
// by the optional URL field.
type Account struct {
	Address [common.AddressLength]byte `json:"address"` // Paradigm account address derived from the key
	URL     URL                        `json:"url"`     // Optional resource locator within a backend
}

// WalletEvent is an event fired by an account backend when a wallet arrival or
// departure is detected.
type WalletEvent struct {
	Wallet Wallet          // Wallet instance arrived or departed
	Kind   WalletEventType // Event type that happened in the system
}

// WalletEventType represents the different event types.
type WalletEventType int

//preset wallet event types for event
const (
	WalletArrived WalletEventType = iota

	// WalletOpened is fired when a wallet is successfully opened with the purpose
	// of starting any background processes such as automatic key derivation.
	WalletOpened

	// WalletDropped
	WalletDropped
)
