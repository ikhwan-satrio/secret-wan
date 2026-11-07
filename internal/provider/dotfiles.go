package provider

import (
	"fmt"
	"os"
)

// DotfilesProvider stores the encrypted payload in the local file configured by
// AppConfig.EncFile. The contents are already age-encrypted.
type DotfilesProvider struct {
	FilePath string
}

func (p *DotfilesProvider) Save(data []byte) error {
	return os.WriteFile(p.FilePath, data, 0644)
}

func (p *DotfilesProvider) Load() ([]byte, error) {
	return os.ReadFile(p.FilePath)
}

// Check verifies that the encrypted file exists and is readable.
func (p *DotfilesProvider) Check() error {
	info, err := os.Stat(p.FilePath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory", p.FilePath)
	}
	if info.Size() == 0 {
		return fmt.Errorf("%s is empty", p.FilePath)
	}

	return nil
}
