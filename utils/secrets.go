package utils

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/ikhwan-satrio/secret-wan/config"
	"golang.org/x/term"
)

type SecretModule struct {
	AppConfig config.AppConfig
}

func (s *SecretModule) GenerateSecureSecret() (string, error) {
	b := make([]byte, s.AppConfig.SecretLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *SecretModule) PromptUserInput(key string, isSecret bool) (string, error) {
	if isSecret {
		fmt.Printf("Masukkan value untuk %s (input tersembunyi): ", key)
		// Gunakan term.ReadPassword agar karakter tidak muncul di layar
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println() // Print newline karena ReadPassword tidak melakukannya
		if err != nil {
			return "", err
		}
		return string(bytePassword), nil
	}

	fmt.Printf("Masukkan value untuk %s: ", key)
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}
