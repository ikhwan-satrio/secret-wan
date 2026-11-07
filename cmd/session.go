package cmd

import (
	"github.com/ikhwan-satrio/secret-wan/internal/app"
	"github.com/ikhwan-satrio/secret-wan/internal/manifest"
)

// cliSession wraps the application service with helpers that translate service
// errors into the CLI's styled error output. Commands use these helpers to keep
// flag parsing and rendering separate from business logic.
type cliSession struct {
	*app.Service
}

func newSession() *cliSession {
	return &cliSession{Service: app.New()}
}

// requireProject parses the manifest and resolves the storage provider.
func (s *cliSession) requireProject() {
	if err := s.EnsureProject(); err != nil {
		failApp("Failed to parse manifest", err)
	}
}

// requireIdentity loads or creates the age identity used for encryption.
func (s *cliSession) requireIdentity() {
	if err := s.EnsureIdentity(); err != nil {
		failApp("Failed to load master key", err)
	}
}

// requireExistingIdentity loads the age identity without creating a new one.
// Read-only commands use this so a missing key is reported instead of silently
// generating a key that cannot decrypt existing storage.
func (s *cliSession) requireExistingIdentity() {
	if err := s.LoadIdentity(); err != nil {
		failApp("Failed to load master key", err)
	}
}

// requireProfile resolves a profile or exits with the available profile list.
func (s *cliSession) requireProfile(name string) *manifest.Profile {
	profile, err := s.Profile(name)
	if err != nil {
		failApp("Invalid profile", err)
	}

	return profile
}

// requireSecrets loads and decrypts stored secrets. A missing store is reported
// with a setup hint instead of an empty result.
func (s *cliSession) requireSecrets() map[string]string {
	secrets, err := s.LoadSecrets()
	if err != nil {
		failApp("Failed to decrypt secrets", err)
	}

	return secrets
}

// loadSecretsAllowMissing loads secrets but treats a missing store as an empty
// map, so `set` can create the first secret.
func (s *cliSession) loadSecretsAllowMissing() map[string]string {
	secrets, missing, err := s.LoadSecretsAllowMissing()
	if err != nil {
		failApp("Failed to decrypt secrets", err)
	}

	if missing {
		warn("No existing secrets found; a new storage will be created.")
	}

	return secrets
}

// saveSecrets encrypts and saves the full map for every configured recipient.
func (s *cliSession) saveSecrets(secrets map[string]string, allowSelfExclusion bool) int {
	count, err := s.SaveSecrets(secrets, allowSelfExclusion)
	if err != nil {
		failApp("Failed to save secrets", err)
	}

	return count
}

// rekey decrypts the current store and writes it back for the configured
// recipients.
func (s *cliSession) rekey(allowSelfExclusion bool) *app.RekeyResult {
	result, err := s.Rekey(allowSelfExclusion)
	if err != nil {
		failApp("Failed to re-encrypt secrets", err)
	}

	return result
}
