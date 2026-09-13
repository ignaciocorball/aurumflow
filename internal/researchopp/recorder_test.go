package researchopp

import (
	"os"
	"testing"
	"time"
)

func TestRecorderImmutableT0AndHash(t *testing.T) {
	dir := t.TempDir()
	t0 := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	s := Signal{ID: "t1", T0: t0, WorldHash: "abc123", Ranking: []byte(`{"rank":1}`), Tier: "IGNORE"}
	if err := Record(dir, s); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(Path(dir))
	if err != nil || len(b) == 0 {
		t.Fatal(err)
	}
	if err := Record(dir, Signal{ID: "bad", WorldHash: ""}); err == nil {
		t.Fatal("empty hash accepted")
	}
}

func TestProspectiveDedupAndVersions(t *testing.T) {
	t0 := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	a := Stamp(Signal{ID: "GOLD", T0: t0, WorldHash: "h1", Tier: "IGNORE"})
	if a.AttentionSpecVersion == "" || a.WorldStateVersion == "" || a.MarketFeaturesVersion == "" || a.GitCommit == "" {
		t.Fatal(a)
	}
	b := Signal{ID: "GOLD", T0: t0.Add(time.Minute), WorldHash: "h1", Tier: "IGNORE"}
	if ShouldRecord(&a, b) {
		t.Fatal("identical snapshot within 5m")
	}
	b.T0 = t0.Add(6 * time.Minute)
	if !ShouldRecord(&a, b) {
		t.Fatal("5m cadence")
	}
	b.T0 = t0.Add(time.Minute)
	b.Tier = "TIER_C"
	if !ShouldRecord(&a, b) {
		t.Fatal("tier change")
	}
	if ShouldRecord(&a, Signal{WorldHash: ""}) {
		t.Fatal("empty hash")
	}
}

func TestOutcomesStaySeparate(t *testing.T) {
	s := Signal{ID: "x", WorldHash: "h", T0: time.Now().UTC()}
	if s.Outcomes != nil {
		t.Fatal("future labels must be empty at t0")
	}
}
