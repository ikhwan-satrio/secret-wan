package provider

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ikhwan-satrio/secret-wan/internal/config"
)

// Factory builds a storage provider for one project.
type Factory func(config.AppConfig, string) StorageProvider

// Descriptor describes one storage provider.
type Descriptor struct {
	Name     string
	ReadOnly bool
	New      Factory
}

var registry = map[string]Descriptor{
	"dotfiles": {
		Name: "dotfiles",
		New: func(appConfig config.AppConfig, _ string) StorageProvider {
			return &DotfilesProvider{FilePath: appConfig.EncFile}
		},
	},
	"keyring": {
		Name: "keyring",
		New: func(_ config.AppConfig, projectName string) StorageProvider {
			return &KeyringProvider{ServiceName: "secret-wan", Key: projectName}
		},
	},
	"docker": {
		Name:     "docker",
		ReadOnly: true,
		New: func(_ config.AppConfig, projectName string) StorageProvider {
			return &DockerSecretProvider{SecretName: projectName}
		},
	},
	"env": {
		Name:     "env",
		ReadOnly: true,
		New:      newEnvProvider,
	},
	"s3": {
		Name: "s3",
		New:  newS3Provider,
	},
}

// Normalize maps an empty provider name to the default and lowercases the rest.
func Normalize(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return "dotfiles"
	}

	return name
}

// Names returns every registered provider name in sorted order.
func Names() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)

	return names
}

// IsKnown reports whether name resolves to a registered provider. An empty name
// is treated as the default provider.
func IsKnown(name string) bool {
	_, exists := registry[Normalize(name)]
	return exists
}

// IsReadOnly reports whether the named provider can only load payloads.
func IsReadOnly(name string) bool {
	descriptor, exists := registry[Normalize(name)]
	return exists && descriptor.ReadOnly
}

// New builds a provider by name. Unknown names return an error.
func New(name string, appConfig config.AppConfig, projectName string) (StorageProvider, error) {
	normalized := Normalize(name)
	descriptor, exists := registry[normalized]
	if !exists {
		return nil, fmt.Errorf("unknown provider %q (expected one of: %s)",
			name, strings.Join(Names(), ", "))
	}

	return descriptor.New(appConfig, projectName), nil
}

// GetStorageProvider returns the provider for a manifest value. Unknown names
// fall back to dotfiles for backward compatibility; `check --strict` and
// `doctor` are responsible for reporting the misconfiguration.
func GetStorageProvider(appConfig config.AppConfig, providerName string, projectName string) StorageProvider {
	instance, err := New(providerName, appConfig, projectName)
	if err != nil {
		return &DotfilesProvider{FilePath: appConfig.EncFile}
	}

	return instance
}
