package provider

import (
	"reflect"
	"testing"

	"github.com/ikhwan-satrio/secret-wan/internal/config"
)

// testEncFile stands in for AppConfig.EncFile so the tests prove the configured
// name is honored instead of a hardcoded default.
const testEncFile = "secrets.enc"

func TestNamesAreSortedAndComplete(t *testing.T) {
	got := Names()
	want := []string{"docker", "dotfiles", "env", "keyring", "s3"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Names() = %v, want %v", got, want)
	}
}

func TestIsKnownNormalizesNames(t *testing.T) {
	for _, name := range []string{"", "dotfiles", " DOTFILES ", "keyring", "s3", "env", "docker"} {
		if !IsKnown(name) {
			t.Errorf("IsKnown(%q) = false, want true", name)
		}
	}

	if IsKnown("vault") {
		t.Error("IsKnown(vault) = true, want false")
	}
}

func TestIsReadOnly(t *testing.T) {
	if !IsReadOnly("docker") {
		t.Error("IsReadOnly(docker) = false, want true")
	}
	if !IsReadOnly("env") {
		t.Error("IsReadOnly(env) = false, want true")
	}
	if IsReadOnly("s3") {
		t.Error("IsReadOnly(s3) = true, want false")
	}
	if IsReadOnly("unknown") {
		t.Error("IsReadOnly(unknown) = true, want false")
	}
}

func TestNewBuildsKnownProviders(t *testing.T) {
	appConfig := config.AppConfig{EncFile: testEncFile}

	cases := map[string]any{
		"dotfiles": &DotfilesProvider{},
		"keyring":  &KeyringProvider{},
		"docker":   &DockerSecretProvider{},
		"env":      &EnvProvider{},
		"s3":       &S3Provider{},
	}

	for name, wantType := range cases {
		instance, err := New(name, appConfig, "demo")
		if err != nil {
			t.Fatalf("New(%q) error = %v", name, err)
		}
		if reflect.TypeOf(instance) != reflect.TypeOf(wantType) {
			t.Errorf("New(%q) = %T, want %T", name, instance, wantType)
		}
	}
}

func TestNewRejectsUnknownProvider(t *testing.T) {
	if _, err := New("vault", config.AppConfig{}, "demo"); err == nil {
		t.Fatal("New(vault) error = nil, want an error")
	}
}

func TestGetStorageProviderFallsBackToDotfiles(t *testing.T) {
	appConfig := config.AppConfig{EncFile: testEncFile}

	instance := GetStorageProvider(appConfig, "vault", "demo")
	dotfiles, ok := instance.(*DotfilesProvider)
	if !ok {
		t.Fatalf("GetStorageProvider(vault) = %T, want *DotfilesProvider", instance)
	}
	if dotfiles.FilePath != testEncFile {
		t.Fatalf("FilePath = %q, want %q", dotfiles.FilePath, testEncFile)
	}
}
