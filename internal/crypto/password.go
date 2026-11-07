package crypto

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const passwordCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>?"

// GenerateSecurePassword returns a random password of the requested length.
// Indices are drawn with crypto/rand.Int, which avoids modulo bias.
func (s *CryptoModule) GenerateSecurePassword(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("password length must be greater than 0")
	}

	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(passwordCharset)))

	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		result[i] = passwordCharset[idx.Int64()]
	}

	return string(result), nil
}
