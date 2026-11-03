package osutil

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// OS merepresentasikan nilai runtime.GOOS yang didukung.
// Gunakan konstanta di bawah ini agar tidak perlu mengingat nama
// internal Go (contoh: macOS adalah "darwin").
type OS string

const (
	Linux   OS = "linux"
	Windows OS = "windows"
	MacOS   OS = "darwin"
	FreeBSD OS = "freebsd"
	OpenBSD OS = "openbsd"
	NetBSD  OS = "netbsd"
	Android OS = "android"
	IOS     OS = "ios"
)

// CurrentOS mengembalikan sistem operasi tempat program berjalan.
func CurrentOS() OS {
	return OS(runtime.GOOS)
}

// IsWindows mengembalikan true jika program berjalan di Windows.
func IsWindows() bool {
	return CurrentOS() == Windows
}

// IsLinux mengembalikan true jika program berjalan di Linux.
func IsLinux() bool {
	return CurrentOS() == Linux
}

// IsMacOS mengembalikan true jika program berjalan di macOS.
func IsMacOS() bool {
	return CurrentOS() == MacOS
}

// GetByOS mengembalikan value yang terdaftar untuk OS saat ini.
// Jika OS saat ini tidak ada di dalam map, fallback akan dikembalikan.
//
// Contoh:
//
//	configPath := osutil.GetByOS(map[osutil.OS]string{
//		osutil.Windows: `C:\Users\me\.config\app\keys.txt`,
//	}, "/home/me/.config/app/keys.txt")
func GetByOS[T any](values map[OS]T, fallback T) T {
	if value, ok := values[CurrentOS()]; ok {
		return value
	}

	return fallback
}

// HomeDir mengembalikan direktori home milik user saat ini.
func HomeDir() (string, error) {
	return os.UserHomeDir()
}

// ConfigDir mengembalikan direktori config standar untuk aplikasi.
// Lokasinya mengikuti aturan OS:
//   - Windows: %AppData%\<appName>
//   - macOS: ~/Library/Application Support/<appName>
//   - Linux/Unix: $XDG_CONFIG_HOME/<appName> atau ~/.config/<appName>
func ConfigDir(appName string) (string, error) {
	if appName == "" {
		return "", errors.New("app name must not be empty")
	}

	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(baseDir, appName), nil
}

// ConfigPath mengembalikan path file di dalam direktori config aplikasi.
func ConfigPath(appName, fileName string) (string, error) {
	if fileName == "" {
		return "", errors.New("file name must not be empty")
	}

	dir, err := ConfigDir(appName)
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, fileName), nil
}

// DockerSecretsDir mengembalikan lokasi default mount Docker secrets
// sesuai sistem operasi.
func DockerSecretsDir() string {
	return GetByOS(map[OS]string{
		Windows: `C:\ProgramData\Docker\secrets`,
	}, "/run/secrets")
}
