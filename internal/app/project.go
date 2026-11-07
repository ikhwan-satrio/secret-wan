package app

import (
	"fmt"
	"os"

	"github.com/ikhwan-satrio/secret-wan/internal/manifest"
	"github.com/ikhwan-satrio/secret-wan/internal/provider"
)

// LoadProject parses the manifest and resolves the storage provider.
func (s *Service) LoadProject() error {
	parsed, err := s.Parser.ParseManifest()
	if err != nil {
		return err
	}

	s.Manifest = parsed
	s.Storage = provider.GetStorageProvider(s.Config, parsed.Project.Provider, parsed.Project.Name)

	return nil
}

// EnsureProject loads the manifest once. It is safe to call repeatedly.
func (s *Service) EnsureProject() error {
	if s.Manifest != nil && s.Storage != nil {
		return nil
	}

	return s.LoadProject()
}

// EnsureIdentity loads an existing key or creates one when missing.
func (s *Service) EnsureIdentity() error {
	if s.Identity != nil {
		return nil
	}

	identity, err := s.Crypto.LoadOrGenerateKey()
	if err != nil {
		return err
	}

	s.Identity = identity

	return nil
}

// LoadIdentity loads an existing key without creating a new one.
func (s *Service) LoadIdentity() error {
	if s.Identity != nil {
		return nil
	}

	identity, err := s.Crypto.LoadKey()
	if err != nil {
		if os.IsNotExist(err) {
			return newUserError("Master key not found", "Run 'secwan init' or 'secwan set' first.")
		}
		return err
	}

	s.Identity = identity

	return nil
}

// Profile returns a named profile or a user-facing error listing alternatives.
func (s *Service) Profile(name string) (*manifest.Profile, error) {
	if err := s.EnsureProject(); err != nil {
		return nil, err
	}

	profile, exists := s.Manifest.Profiles[name]
	if !exists {
		return nil, newUserError("Invalid profile", fmt.Sprintf(
			"Profile '%s' not found. Available profiles: %v",
			name, manifest.GetProfileNames(s.Manifest.Profiles),
		))
	}

	return profile, nil
}
