package provider

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/99designs/keyring"
	"github.com/ikhwan-satrio/secret-wan/internal/config"
	"github.com/ikhwan-satrio/secret-wan/internal/osutil"
)

// IsNotFound reports whether the storage error means "nothing has been stored
// yet" (as opposed to a corrupt or unreadable store).
func IsNotFound(err error) bool {
	return errors.Is(err, os.ErrNotExist) || errors.Is(err, keyring.ErrKeyNotFound)
}

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

// Check verifies that the encrypted file exists and is readable.
func (p *DotfilesProvider) Check() error {
	info, err := os.Stat(p.FilePath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory", p.FilePath)
	}
	if info.Size() == 0 {
		return fmt.Errorf("%s is empty", p.FilePath)
	}

	return nil
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

// Check verifies that a keyring backend is available on this system.
func (p *KeyringProvider) Check() error {
	_, err := keyring.Open(keyring.Config{
		ServiceName: p.ServiceName,
	})
	return err
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

// Check verifies that the Docker secret is mounted and readable.
func (p *DockerSecretProvider) Check() error {
	filePath := filepath.Join(osutil.DockerSecretsDir(), p.SecretName)

	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory", filePath)
	}

	return nil
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
