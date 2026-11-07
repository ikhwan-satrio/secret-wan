package provider

import (
	"encoding/base64"
	"errors"
	"testing"
)

func TestEnvProviderLoadsBase64(t *testing.T) {
	payload := []byte("age-encrypted-bytes")
	t.Setenv(DefaultEnvVar, base64.StdEncoding.EncodeToString(payload))

	provider := &EnvProvider{}
	got, err := provider.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if string(got) != string(payload) {
		t.Fatalf("Load() = %q, want %q", got, payload)
	}

	if err := provider.Check(); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestEnvProviderMissingVariableIsNotFound(t *testing.T) {
	t.Setenv(DefaultEnvVar, "")

	_, err := (&EnvProvider{}).Load()
	if !IsNotFound(err) {
		t.Fatalf("Load() error = %v, want ErrNotFound", err)
	}
}

func TestEnvProviderRejectsInvalidBase64(t *testing.T) {
	t.Setenv(DefaultEnvVar, "not base64!")

	_, err := (&EnvProvider{}).Load()
	if err == nil {
		t.Fatal("Load() error = nil, want an error")
	}
	if errors.Is(err, ErrNotFound) {
		t.Fatalf("Load() error = %v, want a decode error, not ErrNotFound", err)
	}
}

func TestEnvProviderSaveIsUnsupported(t *testing.T) {
	provider := &EnvProvider{}
	if err := provider.Save([]byte("data")); err == nil {
		t.Fatal("Save() error = nil, want an error")
	}
}

func TestEnvProviderAcceptsURISafeAndRawBase64(t *testing.T) {
	payload := []byte{0xfb, 0xff, 0x00, 0x01}
	encodings := []*base64.Encoding{
		base64.URLEncoding,
		base64.RawURLEncoding,
		base64.RawStdEncoding,
	}

	for _, encoding := range encodings {
		t.Setenv(DefaultEnvVar, encoding.EncodeToString(payload))

		got, err := (&EnvProvider{}).Load()
		if err != nil {
			t.Fatalf("Load() with %v error = %v", encoding, err)
		}
		if string(got) != string(payload) {
			t.Fatalf("Load() with %v = %v, want %v", encoding, got, payload)
		}
	}
}
