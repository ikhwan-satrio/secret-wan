package manifest

import (
	"reflect"
	"testing"
)

func TestProfileKeysDeduplicatesAndSorts(t *testing.T) {
	profile := &Profile{
		Required: []string{"B", "A"},
		AutoGen:  []string{"A", "C"},
	}

	got := profile.Keys()
	want := []string{"A", "B", "C"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Profile.Keys() = %v, want %v", got, want)
	}
}

func TestProfilePromptKeysSkipsGenerated(t *testing.T) {
	profile := &Profile{
		Required: []string{"A", "B"},
		AutoGen:  []string{"B", "C"},
	}

	got := profile.PromptKeys()
	want := []string{"A"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Profile.PromptKeys() = %v, want %v", got, want)
	}
}

func TestManifestDeclaredKeysIncludesEveryProfile(t *testing.T) {
	parsed := &Manifest{
		Profiles: map[string]*Profile{
			"default": {Declared: []string{"A", "B"}},
			"staging": {Declared: []string{"C"}},
		},
	}

	got := parsed.DeclaredKeys()
	for _, key := range []string{"A", "B", "C"} {
		if !got[key] {
			t.Errorf("Manifest.DeclaredKeys() missing %q", key)
		}
	}
	if got["D"] {
		t.Error("Manifest.DeclaredKeys() unexpectedly contains D")
	}
}

func TestProfileHasKeyUsesRulesAndLists(t *testing.T) {
	profile := &Profile{
		Required: []string{"A"},
		AutoGen:  []string{"B"},
		Rules:    map[string]SecretRule{"C": {}},
	}

	for _, key := range []string{"A", "B", "C"} {
		if !profile.HasKey(key) {
			t.Errorf("Profile.HasKey(%q) = false, want true", key)
		}
	}
	if profile.HasKey("D") {
		t.Error("Profile.HasKey(D) = true, want false")
	}
}

func TestManifestProfilesContaining(t *testing.T) {
	parsed := &Manifest{
		Profiles: map[string]*Profile{
			"staging": {Required: []string{"A"}},
			"default": {Declared: []string{"A"}, Rules: map[string]SecretRule{"A": {}}},
		},
	}

	got := parsed.ProfilesContaining("A")
	want := []string{"default", "staging"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Manifest.ProfilesContaining() = %v, want %v", got, want)
	}
}
