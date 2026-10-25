package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetTemplate(provider string) string {
	cwd, _ := os.Getwd()
	parrent := filepath.Base(cwd)

	specTemplate := fmt.Sprintf(`
[project]
name = "%s"
provider = "%s"

[profiles.default]
DB_PASS = { generate = true }
DB_URL = { required = true }
	`, parrent, provider)

	return specTemplate
}
