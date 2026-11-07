package provider

import (
	"github.com/99designs/keyring"
)

// KeyringProvider stores the encrypted payload in the operating system's
// credential store.
type KeyringProvider struct {
	ServiceName string
	Key         string
}

func (p *KeyringProvider) Save(data []byte) error {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: p.ServiceName,
	})
	if err != nil {
		return err
	}
	return ring.Set(keyring.Item{
		Key:  p.Key,
		Data: data,
	})
}

func (p *KeyringProvider) Load() ([]byte, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: p.ServiceName,
	})
	if err != nil {
		return nil, err
	}

	item, err := ring.Get(p.Key)
	if err != nil {
		return nil, err
	}
	return item.Data, nil
}

// Check verifies that a keyring backend is available on this system.
func (p *KeyringProvider) Check() error {
	_, err := keyring.Open(keyring.Config{
		ServiceName: p.ServiceName,
	})
	return err
}
