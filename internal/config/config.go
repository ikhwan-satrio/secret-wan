package config

import (
	"log"

	"github.com/ikhwan-satrio/secret-wan/internal/osutil"
)

type AppConfig struct {
	SpecFile        string
	EncFile         string
	GlobalMasterKey string
	AppName         string
	SecretLength    int
}

func UseApp() AppConfig {
	configPath, err := osutil.ConfigPath("secret-wan", "keys.txt")
	if err != nil {
		log.Fatalln(err)
	}

	return AppConfig{
		SpecFile:        "secret-wan.toml",
		EncFile:         "env.enc",
		GlobalMasterKey: configPath,
		AppName:         "secret-wan",
		SecretLength:    100,
	}
}
