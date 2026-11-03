package manifest

import (
	"fmt"
	"os"
	"sort"

	"github.com/BurntSushi/toml"
	"github.com/ikhwan-satrio/secret-wan/internal/config"
)

type tomlManifest struct {
	Project  ProjectInfo            `toml:"project"`
	Profiles map[string]tomlProfile `toml:"profiles"`
}

type tomlProfile map[string]SecretRule

type SecretRule struct {
	Generate bool `toml:"generate"`
	Required bool `toml:"required"`
}

type ProjectInfo struct {
	Name     string `toml:"name"`
	Provider string `toml:"provider"`
}

type Profile struct {
	Name     string
	Required []string
	AutoGen  []string
}

type Manifest struct {
	Project  ProjectInfo
	Profiles map[string]*Profile
}

type ParserModule struct {
	AppConfig config.AppConfig
}

func (p *ParserModule) ParseManifest() (*Manifest, error) {
	data, err := os.ReadFile(p.AppConfig.SpecFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	var raw tomlManifest
	// Gunakan toml.Decode (mengembalikan MetaData dan error)
	if _, err := toml.Decode(string(data), &raw); err != nil {
		return nil, fmt.Errorf("invalid TOML format: %v", err)
	}

	manifest := &Manifest{
		Project:  raw.Project,
		Profiles: make(map[string]*Profile),
	}

	for profName, profRules := range raw.Profiles {
		profile := &Profile{Name: profName}

		for secretName, rule := range profRules {
			if rule.Generate {
				profile.AutoGen = append(profile.AutoGen, secretName)
			}
			if rule.Required {
				profile.Required = append(profile.Required, secretName)
			}
		}

		sort.Strings(profile.AutoGen)
		sort.Strings(profile.Required)
		manifest.Profiles[profName] = profile
	}

	if len(manifest.Profiles) == 0 {
		return nil, fmt.Errorf("no profiles defined inside 'profiles'")
	}

	return manifest, nil
}
