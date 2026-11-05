package secrets

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ikhwan-satrio/secret-wan/internal/config"
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

// PromptUserInput reads a value from the terminal. Prompts are written to
// stderr so stdout stays clean for scripting. When stdin is not a terminal,
// the value is read from the pipe instead.
func (s *SecretModule) PromptUserInput(key string, isSecret bool) (string, error) {
	if isSecret && !term.IsTerminal(int(os.Stdin.Fd())) {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return strings.TrimRight(string(data), "\r\n"), nil
	}

	if isSecret {
		fmt.Fprintf(os.Stderr, "Enter value for %s (hidden input): ", key)
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", err
		}
		return string(bytePassword), nil
	}

	fmt.Fprintf(os.Stderr, "Enter value for %s: ", key)
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}
