package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/99designs/keyring"
	"github.com/ikhwan-satrio/secret-wan/internal/config"
	"github.com/ikhwan-satrio/secret-wan/internal/osutil"
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

type DockerSecretProvider struct {
	SecretName string
}

func (p *DockerSecretProvider) Save(data []byte) error {
	return fmt.Errorf("saving to docker secret is not supported directly")
}

func (p *DockerSecretProvider) Load() ([]byte, error) {
	filePath := filepath.Join(osutil.DockerSecretsDir(), p.SecretName)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	cleaned := strings.TrimSpace(string(data))
	return []byte(cleaned), nil
}

func GetStorageProvider(appConfig config.AppConfig, providerName string, projectName string) StorageProvider {
	switch providerName {
	case "keyring":
		return &KeyringProvider{
			ServiceName: "secret-wan",
			Key:         projectName,
		}
	case "docker":
		return &DockerSecretProvider{
			SecretName: projectName,
		}
	default: // "dotfiles"
		return &DotfilesProvider{
			FilePath: appConfig.EncFile,
		}
	}
}
