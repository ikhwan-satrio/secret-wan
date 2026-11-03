package crypto

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"filippo.io/age"
	"github.com/ikhwan-satrio/secret-wan/internal/config"
	"github.com/ikhwan-satrio/secret-wan/internal/osutil"
	"github.com/ikhwan-satrio/secret-wan/internal/provider"
)

type CryptoModule struct {
	AppConfig config.AppConfig
}

func GetKeyPath(appName string) (string, error) {
	dir, err := osutil.ConfigDir(appName)
	if err != nil {
		return "", err
	}
	// Pastikan folder ada dengan permission ketat (hanya owner yang bisa akses)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "keys.txt"), nil
}

const passwordCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>?"

// GenerateSecurePassword menghasilkan string acak yang aman dengan panjang yang ditentukan.
func (s *CryptoModule) GenerateSecurePassword(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("password length must be greater than 0")
	}

	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(passwordCharset)))

	for i := 0; i < length; i++ {
		// crypto/rand.Int menghasilkan angka acak yang aman dan tanpa bias
		idx, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		result[i] = passwordCharset[idx.Int64()]
	}

	return string(result), nil
}

// LoadOrGenerateKey memuat key yang sudah ada, atau membuat baru jika belum ada
func (c *CryptoModule) LoadOrGenerateKey() (*age.X25519Identity, error) {
	keyPath, err := GetKeyPath(c.AppConfig.AppName)
	if err != nil {
		return nil, err
	}

	// 1. Coba baca file yang sudah ada
	if _, err := os.Stat(keyPath); err == nil {
		file, err := os.Open(keyPath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		identities, err := age.ParseIdentities(file)
		if err != nil {
			return nil, fmt.Errorf("failed to parse key: %v", err)
		}
		if len(identities) == 0 {
			return nil, fmt.Errorf("key file is empty")
		}

		if id, ok := identities[0].(*age.X25519Identity); ok {
			return id, nil
		}
		return nil, fmt.Errorf("unsupported key type")
	}

	// 2. Jika belum ada, generate key baru
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, err
	}

	// 3. Simpan ke file dengan permission 0600 (hanya owner yang bisa baca/tulis)
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

// SaveSecrets mengenkripsi dan menyimpan secrets menggunakan StorageProvider
func (c *CryptoModule) SaveSecrets(storage provider.StorageProvider, secrets map[string]string, recipient age.Recipient) error {
	var buf bytes.Buffer
	w, err := age.Encrypt(&buf, recipient)
	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return err
	}

	// PERBAIKAN: Gunakan w.Write langsung, jangan convert ke string
	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	return storage.Save(buf.Bytes())
}

// LoadSecrets mendekripsi dan memuat secrets menggunakan StorageProvider
func (c *CryptoModule) LoadSecrets(storage provider.StorageProvider, identity age.Identity) (map[string]string, error) {
	// Load data terenkripsi dari provider
	encryptedData, err := storage.Load()
	if err != nil {
		return nil, err
	}

	// Dekripsi data
	r, err := age.Decrypt(bytes.NewReader(encryptedData), identity)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %v", err)
	}

	// Parse JSON
	var secrets map[string]string
	if err := json.NewDecoder(r).Decode(&secrets); err != nil {
		return nil, fmt.Errorf("failed to parse secret JSON: %v", err)
	}

	return secrets, nil
}
