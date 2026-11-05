package crypto

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"filippo.io/age"
)

type memoryStorage struct {
	data []byte
}

func (m *memoryStorage) Save(data []byte) error {
	m.data = append([]byte(nil), data...)
	return nil
}

func (m *memoryStorage) Load() ([]byte, error) {
	if m.data == nil {
		return nil, os.ErrNotExist
	}

	return m.data, nil
}

func TestSaveSecretsEncryptsForEveryRecipient(t *testing.T) {
	first, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	second, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	outsider, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}

	storage := &memoryStorage{}
	module := CryptoModule{}
	secrets := map[string]string{"API_KEY": "value", "DB_PASS": "secret"}

	if err := module.SaveSecrets(storage, secrets, first.Recipient(), second.Recipient()); err != nil {
		t.Fatalf("SaveSecrets() error = %v", err)
	}

	for name, identity := range map[string]*age.X25519Identity{
		"first":  first,
		"second": second,
	} {
		got, err := module.LoadSecrets(storage, identity)
		if err != nil {
			t.Fatalf("LoadSecrets(%s) error = %v", name, err)
		}
		if !reflect.DeepEqual(got, secrets) {
			t.Fatalf("LoadSecrets(%s) = %v, want %v", name, got, secrets)
		}
	}

	if _, err := module.LoadSecrets(storage, outsider); !errors.Is(err, age.ErrIncorrectIdentity) {
		t.Fatalf("LoadSecrets(outsider) error = %v, want age.ErrIncorrectIdentity", err)
	}
}

func TestSaveSecretsRequiresRecipient(t *testing.T) {
	module := CryptoModule{}
	if err := module.SaveSecrets(&memoryStorage{}, map[string]string{"A": "1"}); err == nil {
		t.Error("SaveSecrets() error = nil, want an error")
	}
}

func TestGenerateSecurePassword(t *testing.T) {
	module := CryptoModule{}

	value, err := module.GenerateSecurePassword(32)
	if err != nil {
		t.Fatalf("GenerateSecurePassword() error = %v", err)
	}
	if len(value) != 32 {
		t.Fatalf("GenerateSecurePassword() length = %d, want 32", len(value))
	}

	if _, err := module.GenerateSecurePassword(0); err == nil {
		t.Error("GenerateSecurePassword(0) error = nil, want an error")
	}
}
