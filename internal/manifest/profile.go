package manifest

import "sort"

// Keys returns every key managed by the profile: the union of required and
// generated keys, sorted and de-duplicated. A key declared as both rules appears
// only once.
func (p *Profile) Keys() []string {
	set := make(map[string]bool, len(p.Required)+len(p.AutoGen))

	for _, key := range p.Required {
		set[key] = true
	}
	for _, key := range p.AutoGen {
		set[key] = true
	}

	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return keys
}

// PromptKeys returns required keys that are not auto-generated. A key with both
// rules should be generated, not prompted twice.
func (p *Profile) PromptKeys() []string {
	generated := make(map[string]bool, len(p.AutoGen))
	for _, key := range p.AutoGen {
		generated[key] = true
	}

	keys := make([]string, 0, len(p.Required))
	for _, key := range p.Required {
		if !generated[key] {
			keys = append(keys, key)
		}
	}

	return keys
}

// HasKey reports whether the profile mentions key in its rules or in either key
// list. The slice fallbacks keep the method safe for profiles constructed
// without a Rules map.
func (p *Profile) HasKey(key string) bool {
	if _, ok := p.Rules[key]; ok {
		return true
	}

	for _, candidate := range p.Required {
		if candidate == key {
			return true
		}
	}
	for _, candidate := range p.AutoGen {
		if candidate == key {
			return true
		}
	}

	return false
}

// Rule returns the key's declared rule, or a zero rule when the profile has no
// rule for it. This avoids nil-map panics in checks and reports.
func (p *Profile) Rule(key string) SecretRule {
	if p.Rules == nil {
		return SecretRule{}
	}

	return p.Rules[key]
}

// AllKeys returns the union of the profile's managed keys and every key present
// in stored, sorted alphabetically. It backs `list --all`.
func (p *Profile) AllKeys(stored map[string]string) []string {
	set := make(map[string]bool, len(stored)+len(p.Required)+len(p.AutoGen))

	for _, key := range p.Required {
		set[key] = true
	}
	for _, key := range p.AutoGen {
		set[key] = true
	}
	for key := range stored {
		set[key] = true
	}

	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return keys
}

// DeclaredKeys collects every key mentioned by any profile. Stored keys outside
// this set are considered unmanaged.
func (m *Manifest) DeclaredKeys() map[string]bool {
	declared := make(map[string]bool)

	for _, profile := range m.Profiles {
		for _, key := range profile.Declared {
			declared[key] = true
		}
		// Defensive fallback for profiles constructed without Declared.
		for _, key := range profile.Required {
			declared[key] = true
		}
		for _, key := range profile.AutoGen {
			declared[key] = true
		}
	}

	return declared
}

// ProfilesContaining returns the sorted names of profiles that declare key.
func (m *Manifest) ProfilesContaining(key string) []string {
	var names []string
	for name, profile := range m.Profiles {
		if profile.HasKey(key) {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	return names
}
