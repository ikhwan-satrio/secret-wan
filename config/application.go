package config

import (
	"fmt"
	"log"
	"os"
)

type AppConfig struct {
	SpecFile        string
	EncFile         string
	GlobalMasterKey string
	AppName         string
	SecretLength    int
}

func UseApp() AppConfig {
	home, err := os.UserHomeDir()

	if err != nil {
		log.Fatalln(err)
	}

	config := fmt.Sprintf("%s/.config/secret-wan/keys.txt", home)

	return AppConfig{
		SpecFile:        "secret-wan.toml",
		EncFile:         "env.enc",
		GlobalMasterKey: config,
		AppName:         "secret-wan",
		SecretLength:    100,
	}
}
