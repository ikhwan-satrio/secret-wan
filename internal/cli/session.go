// Package cli contains the runtime shared by command implementations: a
// session that wraps the application service with styled failure helpers, and
// the output helpers themselves.
//
// Commands in the cmd package stay thin by delegating flag parsing to cobra,
// state to app.Service, and error rendering to this package.
package cli

import (
	"github.com/ikhwan-satrio/secret-wan/internal/app"
	"github.com/ikhwan-satrio/secret-wan/internal/manifest"
)

// Session wraps the application service with helpers that translate service
// errors into the CLI's styled error output.
type Session struct {
	*app.Service
}

// NewSession builds a session from the standard application configuration.
func NewSession() *Session {
	return &Session{Service: app.New()}
}

// RequireProject parses the manifest and resolves the storage provider.
func (s *Session) RequireProject() {
	if err := s.EnsureProject(); err != nil {
		FailApp("Failed to parse manifest", err)
	}
}

// RequireIdentity loads or creates the age identity used for encryption.
func (s *Session) RequireIdentity() {
	if err := s.EnsureIdentity(); err != nil {
		FailApp("Failed to load master key", err)
	}
}

// RequireExistingIdentity loads the age identity without creating a new one.
// Read-only commands use this so a missing key is reported instead of silently
// generating a key that cannot decrypt existing storage.
func (s *Session) RequireExistingIdentity() {
	if err := s.LoadIdentity(); err != nil {
		FailApp("Failed to load master key", err)
	}
}

// RequireProfile resolves a profile or exits with the available profile list.
func (s *Session) RequireProfile(name string) *manifest.Profile {
	profile, err := s.Profile(name)
	if err != nil {
		FailApp("Invalid profile", err)
	}

	return profile
}

// RequireSecrets loads and decrypts stored secrets. A missing store is reported
// with a setup hint instead of an empty result.
func (s *Session) RequireSecrets() map[string]string {
	secrets, err := s.LoadSecrets()
	if err != nil {
		FailApp("Failed to decrypt secrets", err)
	}

	return secrets
}

// LoadSecretsAllowMissing loads secrets but treats a missing store as an empty
// map, reporting whether a new storage will be created.
func (s *Session) LoadSecretsAllowMissing() (map[string]string, bool) {
	secrets, missing, err := s.Service.LoadSecretsAllowMissing()
	if err != nil {
		FailApp("Failed to decrypt secrets", err)
	}

	return secrets, missing
}

// SaveSecrets encrypts and saves the full map for every configured recipient.
func (s *Session) SaveSecrets(secrets map[string]string, allowSelfExclusion bool) int {
	count, err := s.Service.SaveSecrets(secrets, allowSelfExclusion)
	if err != nil {
		FailApp("Failed to save secrets", err)
	}

	return count
}
