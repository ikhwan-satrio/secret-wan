package provider

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/ikhwan-satrio/secret-wan/internal/config"
)

// DefaultEnvVar is the environment variable read by the env provider.
const DefaultEnvVar = "SECWAN_ENV_ENC"

// EnvProvider reads an age-encrypted payload from an environment variable. The
// value must be base64 because environment variables cannot safely carry raw
// ciphertext.
//
// This provider is read-only and is intended for CI pipelines and containers.
type EnvProvider struct {
	VarName string
}

func newEnvProvider(_ config.AppConfig, _ string) StorageProvider {
	return &EnvProvider{VarName: DefaultEnvVar}
}

func (p *EnvProvider) variable() string {
	if strings.TrimSpace(p.VarName) == "" {
		return DefaultEnvVar
	}

	return p.VarName
}

// Save is unsupported because the payload must be provided by the environment.
func (p *EnvProvider) Save(data []byte) error {
	return fmt.Errorf("saving to the env provider is not supported; set %s instead", p.variable())
}

// Load decodes the base64 environment value into ciphertext bytes.
func (p *EnvProvider) Load() ([]byte, error) {
	name := p.variable()
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return nil, fmt.Errorf("%w: %s is not set", ErrNotFound, name)
	}

	data, err := decodeBase64(raw)
	if err != nil {
		return nil, fmt.Errorf("%s is not valid base64: %w", name, err)
	}

	return data, nil
}

// Check verifies that the variable is present and decodable.
func (p *EnvProvider) Check() error {
	_, err := p.Load()
	return err
}

// decodeBase64 accepts the standard and URL-safe alphabets, with or without
// padding. CI systems often strip or re-encode padding, so accepting all forms
// avoids a class of avoidable setup failures.
func decodeBase64(value string) ([]byte, error) {
	encodings := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}

	var lastErr error
	for _, encoding := range encodings {
		data, err := encoding.DecodeString(value)
		if err == nil {
			return data, nil
		}
		lastErr = err
	}

	return nil, lastErr
}
