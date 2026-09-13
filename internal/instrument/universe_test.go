package instrument

import (
	"path/filepath"
	"testing"

	"aurumflow/internal/worlddomain"
)

func TestMappingsPersistAndGuess(t *testing.T) {
	id, reg, ac := GuessCanonical("Gold", "GOLD")
	if id != "GOLD" || ac != worlddomain.AssetPrecious {
		t.Fatal(id, reg, ac)
	}
	dir := t.TempDir()
	p := filepath.Join(dir, "map.json")
	if err := PersistMappings(p, []Mapping{{Canonical: "GOLD", Epic: "GOLD"}}); err != nil {
		t.Fatal(err)
	}
	got, err := LoadMappings(p)
	if err != nil || len(got) != 1 || got[0].Epic != "GOLD" {
		t.Fatal(got, err)
	}
}

func TestNeverInventEpic(t *testing.T) {
	id, _, _ := GuessCanonical("Unknown Widget", "XYZ123")
	if id != "" {
		t.Fatal(id)
	}
}
