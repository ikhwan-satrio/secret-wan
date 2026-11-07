package crypto

import (
	"bytes"
	"encoding/json"
	"fmt"

	"filippo.io/age"
	"github.com/ikhwan-satrio/secret-wan/internal/provider"
)

// SaveSecrets serializes the secret map to JSON, encrypts it for every supplied
// recipient, and hands the ciphertext to storage. The payload contains one age
// recipient stanza per key, so any matching private key can decrypt it.
func (c *CryptoModule) SaveSecrets(storage provider.StorageProvider, secrets map[string]string, recipients ...age.Recipient) error {
	if len(recipients) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}

	var buf bytes.Buffer
	w, err := age.Encrypt(&buf, recipients...)
	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return err
	}

	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	return storage.Save(buf.Bytes())
}

// LoadSecrets reads encrypted bytes from storage, decrypts them with identity,
// and parses the JSON secret map.
func (c *CryptoModule) LoadSecrets(storage provider.StorageProvider, identity age.Identity) (map[string]string, error) {
	encryptedData, err := storage.Load()
	if err != nil {
		return nil, err
	}

	r, err := age.Decrypt(bytes.NewReader(encryptedData), identity)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	var secrets map[string]string
	if err := json.NewDecoder(r).Decode(&secrets); err != nil {
		return nil, fmt.Errorf("failed to parse secret JSON: %w", err)
	}

	return secrets, nil
}
