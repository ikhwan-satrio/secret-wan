package app

import (
	"github.com/ikhwan-satrio/secret-wan/internal/provider"
)

// isNotFound reports whether err means that no secrets have been stored yet.
func isNotFound(err error) bool {
	return provider.IsNotFound(err)
}

// GenerateSecret returns a cryptographically secure value using the configured
// secret length.
func (s *Service) GenerateSecret() (string, error) {
	return s.Crypto.GenerateSecurePassword(s.Config.SecretLength)
}

// LoadSecrets loads and decrypts the stored map. A missing store is reported as
// a user-facing setup error.
func (s *Service) LoadSecrets() (map[string]string, error) {
	secrets, _, err := s.loadSecrets(false)
	return secrets, err
}

// LoadSecretsAllowMissing behaves like LoadSecrets but reports a missing store
// through the boolean instead of returning an error. This lets `set` create the
// first secret.
func (s *Service) LoadSecretsAllowMissing() (map[string]string, bool, error) {
	return s.loadSecrets(true)
}

func (s *Service) loadSecrets(allowMissing bool) (map[string]string, bool, error) {
	if s.Storage == nil {
		return nil, false, newUserError("Project is not loaded", "Run the command from a directory containing secret-wan.toml.")
	}
	if s.Identity == nil {
		return nil, false, newUserError("Master key is not loaded", "Run 'secwan init' or 'secwan set' first.")
	}

	secrets, err := s.Crypto.LoadSecrets(s.Storage, s.Identity)
	if err != nil {
		if provider.IsNotFound(err) {
			if allowMissing {
				return map[string]string{}, true, nil
			}
			return nil, false, newUserError("Secrets storage not found",
				"Run 'secwan sync' or 'secwan set <KEY>' to create it first.")
		}
		return nil, false, err
	}

	if secrets == nil {
		secrets = map[string]string{}
	}

	return secrets, false, nil
}

// SaveSecrets encrypts and saves the full map for every configured recipient.
// It returns the number of recipients used.
func (s *Service) SaveSecrets(secrets map[string]string, allowSelfExclusion bool) (int, error) {
	recips, err := s.EncryptionRecipients(allowSelfExclusion)
	if err != nil {
		return 0, err
	}

	if err := s.Crypto.SaveSecrets(s.Storage, secrets, recips...); err != nil {
		return 0, err
	}

	return len(recips), nil
}
