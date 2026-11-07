// Package crypto owns the age cryptography boundary: master-key handling,
// password generation, and encryption/decryption of the secret map.
package crypto

import "github.com/ikhwan-satrio/secret-wan/internal/config"

// CryptoModule carries the application configuration needed by key paths and
// generated secret length.
type CryptoModule struct {
	AppConfig config.AppConfig
}
