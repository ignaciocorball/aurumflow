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
	same := Signal{ID: "GOLD", T0: t0.Add(time.Minute), WorldHash: "h2", Tier: "IGNORE"}
	if ShouldRecord(&a, same) {
		t.Fatal("AsOf-only world_hash change must not record within 5m")
	}
}

func TestOpportunityIndependentOfV1(t *testing.T) {
	t0 := time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC)
	prev := Stamp(Signal{ID: "GOLD", T0: t0, WorldHash: "h", Tier: "TIER_B"})
	next := Signal{ID: "GOLD", T0: t0.Add(6 * time.Minute), WorldHash: "h", Tier: "TIER_B"}
	if !ShouldRecord(&prev, next) {
		t.Fatal("5m cadence must not require a V1 signal")
	}
}

func TestOutcomeDoesNotRewriteT0(t *testing.T) {
	dir := t.TempDir()
	t0 := time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC)
	s := Signal{ID: "GOLD-1", T0: t0, WorldHash: "abc"}
	if err := Record(dir, s); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(Path(dir))
	rec := BuildOutcome("GOLD-1", t0, 100, t0.Add(20*time.Minute), []PricePoint{{T: t0.Add(time.Minute), P: 101}, {T: t0.Add(16 * time.Minute), P: 102}})
	if rec.Horizons["15m"].Quality == "" {
		t.Fatal(rec)
	}
	if err := AppendOutcomeOnly(dir, rec); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(Path(dir))
	if string(before) != string(after) {
		t.Fatal("t0 rewritten")
	}
}

func TestLabelMatureDoesNotRewriteT0(t *testing.T) {
	dir := t.TempDir()
	t0 := time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC)
	s := Signal{ID: "GOLD-1", T0: t0, WorldHash: "abc", MarketState: []byte(`{"Mid":100}`), IntegrityStatus: "VALID"}
	if err := Record(dir, s); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(Path(dir))
	n15, _, err := LabelMature(dir, t0.Add(16*time.Minute), map[string][]PricePoint{
		"GOLD": {{T: t0.Add(time.Minute), P: 101}, {T: t0.Add(16 * time.Minute), P: 102}},
	}, nil)
	if err != nil || n15 != 1 {
		t.Fatalf("n15=%d err=%v", n15, err)
	}
	after, _ := os.ReadFile(Path(dir))
	if string(before) != string(after) {
		t.Fatal("t0 rewritten")
	}
}

func TestLabelMatureSkipsPreP88AndEmptyPath(t *testing.T) {
	dir := t.TempDir()
	t0 := time.Date(2026, 9, 13, 20, 0, 0, 0, time.UTC)
	legacy := Signal{ID: "GOLD-OLD", T0: t0, WorldHash: "h"}
	if err := Record(dir, legacy); err != nil {
		t.Fatal(err)
	}
	n15, _, err := LabelMature(dir, t0.Add(20*time.Minute), map[string][]PricePoint{"GOLD": {{T: t0.Add(16 * time.Minute), P: 1}}}, nil)
	if err != nil || n15 != 0 {
		t.Fatalf("pre-P8.8 labeled n15=%d err=%v", n15, err)
	}
}

func TestOutcomesStaySeparate(t *testing.T) {
	s := Signal{ID: "x", WorldHash: "h", T0: time.Now().UTC()}
	if s.Outcomes != nil {
		t.Fatal("future labels must be empty at t0")
	}
}
