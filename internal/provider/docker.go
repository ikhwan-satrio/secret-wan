package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ikhwan-satrio/secret-wan/internal/osutil"
)

// DockerSecretProvider reads an encrypted payload from a mounted Docker secret.
// Docker secrets are read-only, so Save always fails.
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
