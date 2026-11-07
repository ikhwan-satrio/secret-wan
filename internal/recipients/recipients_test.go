package recipients

import (
	"path/filepath"
	"reflect"
	"testing"

	"filippo.io/age"
)

func TestParseLinesNormalizesAndSorts(t *testing.T) {
	first, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	second, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}

	keys, err := ParseLines([]string{
		"# project recipients",
		"",
		first.Recipient().String(),
		"  " + second.Recipient().String() + " # teammate",
		first.Recipient().String(),
	})
	if err != nil {
		t.Fatalf("ParseLines() error = %v", err)
	}

	want := []string{first.Recipient().String(), second.Recipient().String()}
	if want[0] > want[1] {
		want[0], want[1] = want[1], want[0]
	}

	if !reflect.DeepEqual(keys, want) {
		t.Fatalf("ParseLines() = %v, want %v", keys, want)
	}
}

func TestParseLinesRejectsInvalidRecipient(t *testing.T) {
	if _, err := ParseLines([]string{"not-a-recipient"}); err == nil {
		t.Error("ParseLines() error = nil, want an error")
	}
}

func TestStoreRoundTrip(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "secret-wan.recipients")
	store := New(path)

	exists, err := store.Exists()
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if exists {
		t.Fatal("Exists() = true before Save")
	}

	if err := store.Save([]string{identity.Recipient().String()}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	exists, err = store.Exists()
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Fatal("Exists() = false after Save")
	}

	keys, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(keys) != 1 || keys[0] != identity.Recipient().String() {
		t.Fatalf("Load() = %v, want the saved recipient", keys)
	}
}

func TestStoreLoadMissingIsEmpty(t *testing.T) {
	store := New(filepath.Join(t.TempDir(), "missing.recipients"))

	keys, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if keys != nil {
		t.Fatalf("Load() = %v, want nil", keys)
	}
}

func TestContainsAndRemove(t *testing.T) {
	keys := []string{"age1", "age2"}

	if !Contains(keys, "age1") {
		t.Error("Contains() = false for existing key")
	}
	if Contains(keys, "age3") {
		t.Error("Contains() = true for missing key")
	}

	got := Remove(keys, "age1")
	if !reflect.DeepEqual(got, []string{"age2"}) {
		t.Fatalf("Remove() = %v, want [age2]", got)
	}
}

func TestNormalizeRejectsEmptyAndInvalid(t *testing.T) {
	if _, err := Normalize([]string{""}); err != nil {
		t.Errorf("Normalize(empty) error = %v, want nil", err)
	}
	if _, err := Normalize([]string{"bad"}); err == nil {
		t.Error("Normalize(invalid) error = nil, want an error")
	}
}

func TestParseRecipientCanonicalizesAndRejectsEmpty(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}

	raw := "  " + identity.Recipient().String() + "  "
	got, err := ParseRecipient(raw)
	if err != nil {
		t.Fatalf("ParseRecipient() error = %v", err)
	}
	if got != identity.Recipient().String() {
		t.Fatalf("ParseRecipient() = %q, want %q", got, identity.Recipient().String())
	}

	if _, err := ParseRecipient("   "); err == nil {
		t.Error("ParseRecipient(empty) error = nil, want an error")
	}
	if _, err := ParseRecipient("bad"); err == nil {
		t.Error("ParseRecipient(invalid) error = nil, want an error")
	}
}
