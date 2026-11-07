package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"
	"github.com/ikhwan-satrio/secret-wan/internal/config"
	"github.com/ikhwan-satrio/secret-wan/internal/manifest"
)

type memoryStorage struct {
	data    []byte
	saveErr error
}

func (m *memoryStorage) Save(data []byte) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.data = append([]byte(nil), data...)
	return nil
}

func (m *memoryStorage) Load() ([]byte, error) {
	if m.data == nil {
		return nil, os.ErrNotExist
	}
	return m.data, nil
}

func newRecipientService(t *testing.T, identity *age.X25519Identity, keys []string) (*Service, *memoryStorage, string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "secret-wan.recipients")
	cfg := config.AppConfig{RecipientsFile: path}
	service := NewWithConfig(cfg)
	service.Identity = identity
	service.Manifest = &manifest.Manifest{
		Project: manifest.ProjectInfo{Name: "demo", Provider: "dotfiles"},
	}
	storage := &memoryStorage{}
	service.Storage = storage

	if len(keys) > 0 {
		if err := service.Recipients.Save(keys); err != nil {
			t.Fatalf("save recipients: %v", err)
		}
	}

	return service, storage, path
}

func generateIdentity(t *testing.T) *age.X25519Identity {
	t.Helper()

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}

	return identity
}

func TestEncryptionRecipientsUsesCurrentKeyWhenNoFile(t *testing.T) {
	self := generateIdentity(t)
	service, _, _ := newRecipientService(t, self, nil)

	got, err := service.EncryptionRecipients(false)
	if err != nil {
		t.Fatalf("EncryptionRecipients() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("EncryptionRecipients() returned %d recipients, want 1", len(got))
	}
}

func TestEncryptionRecipientsRequiresCurrentKey(t *testing.T) {
	self := generateIdentity(t)
	other := generateIdentity(t)
	service, _, _ := newRecipientService(t, self, []string{other.Recipient().String()})

	_, err := service.EncryptionRecipients(false)
	var userErr *UserError
	if !errors.As(err, &userErr) {
		t.Fatalf("EncryptionRecipients() error = %v, want UserError", err)
	}

	got, err := service.EncryptionRecipients(true)
	if err != nil {
		t.Fatalf("EncryptionRecipients(allowSelfExclusion) error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("EncryptionRecipients(allowSelfExclusion) returned %d recipients, want 1", len(got))
	}
}

func TestEncryptionRecipientsKeepsConfiguredKeys(t *testing.T) {
	self := generateIdentity(t)
	other := generateIdentity(t)
	service, _, _ := newRecipientService(t, self, []string{
		self.Recipient().String(),
		other.Recipient().String(),
	})

	got, err := service.EncryptionRecipients(false)
	if err != nil {
		t.Fatalf("EncryptionRecipients() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("EncryptionRecipients() returned %d recipients, want 2", len(got))
	}
}

func TestAddRecipientAllowsSelfReadd(t *testing.T) {
	self := generateIdentity(t)
	other := generateIdentity(t)
	service, _, _ := newRecipientService(t, self, []string{other.Recipient().String()})

	change, err := service.AddRecipient(self.Recipient().String())
	if err != nil {
		t.Fatalf("AddRecipient(self) error = %v", err)
	}
	if change.Secrets != 0 || change.Recipients != 2 {
		t.Fatalf("AddRecipient(self) = %+v, want 0 secrets for 2 recipients", change)
	}

	keys, err := service.Recipients.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !containsAll(keys, self.Recipient().String(), other.Recipient().String()) {
		t.Fatalf("recipient list = %v, want self and other", keys)
	}
}

func TestAddRecipientRefusesDifferentKeyWhenSelfMissing(t *testing.T) {
	self := generateIdentity(t)
	other := generateIdentity(t)
	newcomer := generateIdentity(t)
	service, _, _ := newRecipientService(t, self, []string{other.Recipient().String()})

	if _, err := service.AddRecipient(newcomer.Recipient().String()); err == nil {
		t.Fatal("AddRecipient() error = nil, want a self-lockout error")
	}
}

func TestRemoveRecipientRejectsLastRecipient(t *testing.T) {
	self := generateIdentity(t)
	service, _, _ := newRecipientService(t, self, []string{self.Recipient().String()})

	if _, err := service.RemoveRecipient(self.Recipient().String(), false); err == nil {
		t.Fatal("RemoveRecipient() error = nil, want a last-recipient error")
	}
}

func TestRemoveRecipientRekeysRemainingKey(t *testing.T) {
	self := generateIdentity(t)
	other := generateIdentity(t)
	service, storage, _ := newRecipientService(t, self, []string{
		self.Recipient().String(),
		other.Recipient().String(),
	})

	secrets := map[string]string{"A": "1"}
	if err := service.Crypto.SaveSecrets(storage, secrets, self.Recipient(), other.Recipient()); err != nil {
		t.Fatal(err)
	}

	change, err := service.RemoveRecipient(other.Recipient().String(), false)
	if err != nil {
		t.Fatalf("RemoveRecipient() error = %v", err)
	}
	if change.Secrets != 1 || change.Recipients != 1 {
		t.Fatalf("RemoveRecipient() = %+v, want 1 secret for 1 recipient", change)
	}

	if _, err := service.Crypto.LoadSecrets(storage, self); err != nil {
		t.Fatalf("remaining key cannot decrypt: %v", err)
	}
	if _, err := service.Crypto.LoadSecrets(storage, other); !errors.Is(err, age.ErrIncorrectIdentity) {
		t.Fatalf("removed key decrypt error = %v, want age.ErrIncorrectIdentity", err)
	}
}

func TestAddRecipientRollsBackOnRekeyFailure(t *testing.T) {
	self := generateIdentity(t)
	other := generateIdentity(t)
	service, storage, _ := newRecipientService(t, self, []string{self.Recipient().String()})

	if err := service.Crypto.SaveSecrets(storage, map[string]string{"A": "1"}, self.Recipient()); err != nil {
		t.Fatal(err)
	}

	storage.saveErr = errors.New("disk full")

	_, err := service.AddRecipient(other.Recipient().String())
	var changeErr *RecipientChangeError
	if !errors.As(err, &changeErr) {
		t.Fatalf("AddRecipient() error = %v, want RecipientChangeError", err)
	}
	if !changeErr.Restored {
		t.Fatalf("AddRecipient() rollback restored = false, want true (rollback error: %v)", changeErr.RollbackErr)
	}

	keys, err := service.Recipients.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 || keys[0] != self.Recipient().String() {
		t.Fatalf("recipient list after rollback = %v, want only self", keys)
	}
}

func containsAll(keys []string, wanted ...string) bool {
	for _, key := range wanted {
		found := false
		for _, candidate := range keys {
			if candidate == key {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
