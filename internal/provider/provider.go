// Package provider defines the storage abstraction used for encrypted secret
// payloads and the available storage provider implementations.
//
// Providers move opaque bytes only. Encryption and decryption happen in the
// crypto package, so a provider never sees plaintext.
package provider

import (
	"errors"
	"os"

	"github.com/99designs/keyring"
)

// ErrNotFound marks a storage backend that has no payload yet. Providers wrap
// this error so commands can distinguish a first run from a broken store.
var ErrNotFound = errors.New("secret not found")

// IsNotFound reports whether the storage error means "nothing has been stored
// yet" (as opposed to a corrupt or unreadable store).
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound) ||
		errors.Is(err, os.ErrNotExist) ||
		errors.Is(err, keyring.ErrKeyNotFound)
}

// StorageProvider moves an encrypted payload to and from a storage backend.
type StorageProvider interface {
	Save(data []byte) error
	Load() ([]byte, error)
}
