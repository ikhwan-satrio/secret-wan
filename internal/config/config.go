package config

import (
	"log"

	"github.com/ikhwan-satrio/secret-wan/internal/osutil"
)

type AppConfig struct {
	SpecFile        string
	EncFile         string
	RecipientsFile  string
	GlobalMasterKey string
	AppName         string
	SecretLength    int
}

// RecipientsPath returns the configured recipients file, falling back to the
// default name for callers that constructed AppConfig manually.
func (c AppConfig) RecipientsPath() string {
	if c.RecipientsFile != "" {
		return c.RecipientsFile
	}

	return "secret-wan.recipients"
}

func UseApp() AppConfig {
	configPath, err := osutil.ConfigPath("secret-wan", "keys.txt")
	if err != nil {
		log.Fatalln(err)
	}

	return AppConfig{
		SpecFile:        "secret-wan.toml",
		EncFile:         ".env.enc",
		RecipientsFile:  "secret-wan.recipients",
		GlobalMasterKey: configPath,
		AppName:         "secret-wan",
		SecretLength:    100,
	}
}
