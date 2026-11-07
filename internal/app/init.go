package app

import (
	"fmt"
	"os"

	"github.com/ikhwan-satrio/secret-wan/internal/manifest"
)

// CreateManifest writes the provider-specific starter manifest when the spec
// file does not exist yet. It returns true when a file was created.
func (s *Service) CreateManifest(providerName string) (bool, error) {
	if _, err := os.Stat(s.Config.SpecFile); err == nil {
		return false, nil
	}

	template := manifest.GetTemplate(providerName)
	if err := os.WriteFile(s.Config.SpecFile, []byte(template), 0644); err != nil {
		return false, err
	}

	return true, nil
}

// CreateRecipientList seeds the recipient file with the current key when the
// file does not exist. It returns true when a file was created.
func (s *Service) CreateRecipientList() (bool, error) {
	if s.Identity == nil {
		return false, fmt.Errorf("master key is not loaded")
	}

	exists, err := s.Recipients.Exists()
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}

	if err := s.Recipients.Save([]string{s.Identity.Recipient().String()}); err != nil {
		return false, err
	}

	return true, nil
}
