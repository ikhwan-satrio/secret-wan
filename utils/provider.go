package utils

import (
	"os"

	"github.com/99designs/keyring"
	"github.com/ikhwan-satrio/secret-wan/config"
)

type StorageProvider interface {
	Save(data []byte) error
	Load() ([]byte, error)
}

type DotfilesProvider struct {
	FilePath string
}

func (p *DotfilesProvider) Save(data []byte) error {
	return os.WriteFile(p.FilePath, data, 0644)
}

func (p *DotfilesProvider) Load() ([]byte, error) {
	return os.ReadFile(p.FilePath)
}

type KeyringProvider struct {
	ServiceName string // Nama aplikasi di keychain
	Key         string // Key spesifik (misal: nama project)
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

func GetStorageProvider(appConfig config.AppConfig, providerName string, projectName string) StorageProvider {
	switch providerName {
	case "keyring":
		return &KeyringProvider{
			ServiceName: "secret-wan",
			Key:         projectName,
		}
	default: // "dotfiles"
		return &DotfilesProvider{
			FilePath: appConfig.EncFile,
		}
	}
}
