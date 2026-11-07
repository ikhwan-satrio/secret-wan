package app

import (
	"fmt"

	"filippo.io/age"
	"github.com/ikhwan-satrio/secret-wan/internal/recipients"
)

// RecipientList describes the project's public recipient sidecar and, when
// available, the current user's own recipient.
type RecipientList struct {
	Path      string
	Keys      []string
	Self      string
	SelfKnown bool
}

// SelfListed reports whether the current key appears in a non-empty list.
func (l RecipientList) SelfListed() bool {
	return l.SelfKnown && recipients.Contains(l.Keys, l.Self)
}

// RecipientKeys loads the configured public recipients. A missing file returns
// an empty slice, which means the project still uses single-recipient mode.
func (s *Service) RecipientKeys() ([]string, error) {
	keys, err := s.Recipients.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to read recipients: %w", err)
	}

	return keys, nil
}

// ListRecipients loads the recipient list and tries to identify the current key.
// A missing master key is not fatal because `recipient list` can still display
// the public keys.
func (s *Service) ListRecipients() (RecipientList, error) {
	list := RecipientList{Path: s.Recipients.Path}

	keys, err := s.Recipients.Load()
	if err != nil {
		return list, err
	}
	list.Keys = keys

	if s.Identity != nil {
		list.Self = s.Identity.Recipient().String()
		list.SelfKnown = true
		return list, nil
	}

	identity, keyErr := s.Crypto.LoadKey()
	if keyErr == nil {
		s.Identity = identity
		list.Self = identity.Recipient().String()
		list.SelfKnown = true
	}

	return list, nil
}

// EncryptionRecipients resolves the configured recipient list into age
// recipients and, unless allowSelfExclusion is set, refuses to lock the current
// key out of future writes.
func (s *Service) EncryptionRecipients(allowSelfExclusion bool) ([]age.Recipient, error) {
	if s.Identity == nil {
		return nil, fmt.Errorf("master key is not loaded")
	}

	keys, err := s.Recipients.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to read recipients: %w", err)
	}

	self := s.Identity.Recipient().String()
	if len(keys) == 0 {
		keys = []string{self}
	}

	if !allowSelfExclusion && !recipients.Contains(keys, self) {
		return nil, newUserError("Current key is not a recipient", fmt.Sprintf(
			"current key %s is not a recipient; run 'secwan recipient add %s' first",
			self, self,
		))
	}

	return recipients.Parse(keys)
}

// RekeyResult reports what a rekey operation did.
type RekeyResult struct {
	Secrets      int
	Recipients   int
	UsedImplicit bool
	Missing      bool
	Self         string
	SelfListed   bool
}

// Rekey decrypts the current store and writes it back for the configured
// recipients. A missing store is reported through Missing rather than an error.
func (s *Service) Rekey(allowSelfExclusion bool) (*RekeyResult, error) {
	if s.Identity == nil {
		return nil, fmt.Errorf("master key is not loaded")
	}
	if s.Storage == nil {
		return nil, newUserError("Project is not loaded", "Run the command from a directory containing secret-wan.toml.")
	}

	keys, err := s.Recipients.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to read recipients: %w", err)
	}

	self := s.Identity.Recipient().String()
	result := &RekeyResult{Self: self}
	if len(keys) == 0 {
		keys = []string{self}
		result.UsedImplicit = true
	}
	result.SelfListed = recipients.Contains(keys, self)

	if !allowSelfExclusion && !result.SelfListed {
		return nil, newUserError("Current key is not a recipient", fmt.Sprintf(
			"Add it with 'secwan recipient add %s' or use --force to lock this key out.",
			self,
		))
	}

	secrets, err := s.Crypto.LoadSecrets(s.Storage, s.Identity)
	if err != nil {
		if isNotFound(err) {
			result.Recipients = len(keys)
			result.Missing = true
			return result, nil
		}
		return nil, err
	}

	ageRecipients, err := recipients.Parse(keys)
	if err != nil {
		return nil, newUserError("Invalid recipients", err.Error())
	}

	if err := s.Crypto.SaveSecrets(s.Storage, secrets, ageRecipients...); err != nil {
		return nil, err
	}

	result.Secrets = len(secrets)
	result.Recipients = len(ageRecipients)

	return result, nil
}

// RecipientChange reports the outcome of adding or removing one recipient.
type RecipientChange struct {
	Key            string
	AlreadyPresent bool
	SelfRemoved    bool
	Secrets        int
	Recipients     int
}

// AddRecipient validates a new key, appends it to the list, and rekeys. The
// previous list is restored if rekeying fails. Adding the current key itself is
// the supported recovery path when a list accidentally omits it.
func (s *Service) AddRecipient(raw string) (*RecipientChange, error) {
	newKey, err := recipients.ParseRecipient(raw)
	if err != nil {
		return nil, newUserError("Invalid recipient", err.Error())
	}

	if err := s.EnsureProject(); err != nil {
		return nil, err
	}
	if err := s.EnsureIdentity(); err != nil {
		return nil, err
	}

	keys, err := s.Recipients.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to read recipients: %w", err)
	}

	self := s.Identity.Recipient().String()
	if len(keys) == 0 {
		keys = []string{self}
	}

	if !recipients.Contains(keys, self) && newKey != self {
		return nil, newUserError("Current key is not a recipient", fmt.Sprintf(
			"Add your own key first: secwan recipient add %s", self,
		))
	}

	change := &RecipientChange{Key: newKey}
	if recipients.Contains(keys, newKey) {
		change.AlreadyPresent = true
		return change, nil
	}

	oldKeys := append([]string(nil), keys...)
	newKeys := append(append([]string(nil), keys...), newKey)
	if err := s.Recipients.Save(newKeys); err != nil {
		return nil, err
	}

	result, err := s.Rekey(false)
	if err != nil {
		rollbackErr := s.Recipients.Save(oldKeys)
		return nil, &RecipientChangeError{Err: err, Restored: rollbackErr == nil, RollbackErr: rollbackErr}
	}

	change.Secrets = result.Secrets
	change.Recipients = result.Recipients

	return change, nil
}

// RemoveRecipient validates the target, removes it, and rekeys. Removing the
// current key or a key from a list that already excludes it requires force.
func (s *Service) RemoveRecipient(raw string, force bool) (*RecipientChange, error) {
	target, err := recipients.ParseRecipient(raw)
	if err != nil {
		return nil, newUserError("Invalid recipient", err.Error())
	}

	if err := s.EnsureProject(); err != nil {
		return nil, err
	}
	if err := s.LoadIdentity(); err != nil {
		return nil, err
	}

	keys, err := s.Recipients.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to read recipients: %w", err)
	}

	self := s.Identity.Recipient().String()
	if len(keys) == 0 {
		keys = []string{self}
	}

	if !recipients.Contains(keys, target) {
		return nil, newUserError("Recipient not found", fmt.Sprintf("%s is not in %s.", target, s.Recipients.Path))
	}
	if len(keys) == 1 {
		return nil, newUserError("Cannot remove the last recipient", "At least one recipient must remain.")
	}
	if !recipients.Contains(keys, self) && !force {
		return nil, newUserError("Current key is not a recipient", "Use --force only if you intend to lose access to future writes.")
	}

	selfRemoved := target == self
	if selfRemoved && !force {
		return nil, newUserError("Refusing to remove your own key",
			"Use --force to re-encrypt for the remaining recipients. This key will no longer decrypt new writes.")
	}

	oldKeys := append([]string(nil), keys...)
	newKeys := recipients.Remove(keys, target)
	if err := s.Recipients.Save(newKeys); err != nil {
		return nil, err
	}

	result, err := s.Rekey(true)
	if err != nil {
		rollbackErr := s.Recipients.Save(oldKeys)
		return nil, &RecipientChangeError{Err: err, Restored: rollbackErr == nil, RollbackErr: rollbackErr}
	}

	return &RecipientChange{
		Key:         target,
		SelfRemoved: selfRemoved,
		Secrets:     result.Secrets,
		Recipients:  result.Recipients,
	}, nil
}
