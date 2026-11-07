// Package app contains the application service that coordinates
// configuration, cryptography, manifests, storage, and recipient policy.
//
// The cmd package is intentionally thin: it parses flags, calls this service,
// and renders output. Business rules such as self-lockout prevention, recipient
// fallback, and rekey rollback live here.
package app

import (
	"filippo.io/age"
	"github.com/ikhwan-satrio/secret-wan/internal/config"
	"github.com/ikhwan-satrio/secret-wan/internal/crypto"
	"github.com/ikhwan-satrio/secret-wan/internal/manifest"
	"github.com/ikhwan-satrio/secret-wan/internal/provider"
	"github.com/ikhwan-satrio/secret-wan/internal/recipients"
)

// Service holds the loaded project state used by one CLI invocation.
//
// Fields are populated lazily by LoadProject, EnsureIdentity, and LoadIdentity.
type Service struct {
	Config     config.AppConfig
	Crypto     crypto.CryptoModule
	Parser     manifest.ParserModule
	Recipients recipients.Store

	Manifest *manifest.Manifest
	Identity *age.X25519Identity
	Storage  provider.StorageProvider
}

// New builds a service from the standard application configuration.
func New() *Service {
	return NewWithConfig(config.UseApp())
}

// NewWithConfig builds a service with an explicit configuration. Tests use this
// to point the service at temporary files.
func NewWithConfig(appConfig config.AppConfig) *Service {
	return &Service{
		Config:     appConfig,
		Crypto:     crypto.CryptoModule{AppConfig: appConfig},
		Parser:     manifest.ParserModule{AppConfig: appConfig},
		Recipients: recipients.New(appConfig.RecipientsPath()),
	}
}
