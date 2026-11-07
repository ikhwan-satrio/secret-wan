package crypto

import (
	"fmt"
	"os"
	"path/filepath"

	"filippo.io/age"
	"github.com/ikhwan-satrio/secret-wan/internal/osutil"
)

// KeyPath returns the expected location of the master key without creating any
// directories. Useful for diagnostics such as `secwan doctor`.
func KeyPath(appName string) (string, error) {
	dir, err := osutil.ConfigDir(appName)
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "keys.txt"), nil
}

// GetKeyPath returns the master key path and creates the config directory with
// owner-only permissions.
func GetKeyPath(appName string) (string, error) {
	dir, err := osutil.ConfigDir(appName)
	if err != nil {
		return "", err
	}
	// Ensure the directory exists with strict permissions.
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "keys.txt"), nil
}

// LoadKey loads an existing master key. It never creates a new one, which is
// what read-only commands should use.
func (c *CryptoModule) LoadKey() (*age.X25519Identity, error) {
	keyPath, err := KeyPath(c.AppConfig.AppName)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(keyPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	identities, err := age.ParseIdentities(file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse key: %w", err)
	}
	if len(identities) == 0 {
		return nil, fmt.Errorf("key file is empty")
	}

	if id, ok := identities[0].(*age.X25519Identity); ok {
		return id, nil
	}
	return nil, fmt.Errorf("unsupported key type")
}

// LoadOrGenerateKey loads an existing key or creates a new master key when none
// exists yet.
func (c *CryptoModule) LoadOrGenerateKey() (*age.X25519Identity, error) {
	keyPath, err := GetKeyPath(c.AppConfig.AppName)
	if err != nil {
		return nil, err
	}

	// 1. Try to read an existing key file.
	if _, err := os.Stat(keyPath); err == nil {
		return c.LoadKey()
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	// 2. Generate a new identity when the file is missing.
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, err
	}

	// 3. Store the private identity with owner-only permissions.
	file, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if _, err := file.WriteString(identity.String() + "\n"); err != nil {
		return nil, err
	}
	if _, err := file.WriteString("# Public key (share this with your team): " + identity.Recipient().String() + "\n"); err != nil {
		return nil, err
	}

	return identity, nil
}
