package osutil

import (
	"path/filepath"
	"testing"
)

func TestGetByOS(t *testing.T) {
	t.Run("returns value for current OS", func(t *testing.T) {
		values := map[OS]string{CurrentOS(): "native"}

		if got := GetByOS(values, "fallback"); got != "native" {
			t.Errorf("GetByOS() = %q, want %q", got, "native")
		}
	})

	t.Run("returns fallback when OS is not registered", func(t *testing.T) {
		values := map[OS]string{OS("unknown-os"): "other"}

		if got := GetByOS(values, "fallback"); got != "fallback" {
			t.Errorf("GetByOS() = %q, want %q", got, "fallback")
		}
	})
}

func TestConfigPath(t *testing.T) {
	got, err := ConfigPath("secret-wan", "keys.txt")
	if err != nil {
		t.Fatalf("ConfigPath() error = %v", err)
	}

	if filepath.Base(got) != "keys.txt" {
		t.Errorf("ConfigPath() = %q, want file name %q", got, "keys.txt")
	}
	if filepath.Base(filepath.Dir(got)) != "secret-wan" {
		t.Errorf("ConfigPath() = %q, want parent directory %q", got, "secret-wan")
	}
}

func TestConfigPathValidation(t *testing.T) {
	t.Run("rejects empty app name", func(t *testing.T) {
		if _, err := ConfigPath("", "keys.txt"); err == nil {
			t.Error("ConfigPath() error = nil, want an error")
		}
	})

	t.Run("rejects empty file name", func(t *testing.T) {
		if _, err := ConfigPath("secret-wan", ""); err == nil {
			t.Error("ConfigPath() error = nil, want an error")
		}
	})
}

func TestDockerSecretsDir(t *testing.T) {
	want := "/run/secrets"
	if IsWindows() {
		want = `C:\ProgramData\Docker\secrets`
	}

	if got := DockerSecretsDir(); got != want {
		t.Errorf("DockerSecretsDir() = %q, want %q", got, want)
	}
}
