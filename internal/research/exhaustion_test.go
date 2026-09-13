package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"aurumflow/internal/exhaustion"
)

func TestV1ParityWithExhaustionPackage(t *testing.T) {
	cases := [][2]int{{1, -20}, {1, 20}, {-1, 20}, {-1, -20}, {1, 0}, {0, 40}}
	for _, c := range cases {
		a := ClassifyFlow(c[0], float64(c[1]), 15)
		b := exhaustion.ClassifyV1(c[0], float64(c[1]))
		if a != b {
			t.Fatalf("%v %s vs %s", c, a, b)
		}
	}
}

func TestClassifyFlowBoundaries(t *testing.T) {
	if ClassifyFlow(1, -15, 15) != FlowExhaustion {
		t.Fatal("LONG negative 15")
	}
	if ClassifyFlow(1, -14.9, 15) != FlowNeutral {
		t.Fatal("LONG just inside")
	}
	if ClassifyFlow(-1, 15, 15) != FlowExhaustion {
		t.Fatal("SHORT positive 15")
	}
	if ClassifyFlow(1, 15, 15) != FlowContinuation {
		t.Fatal("LONG positive")
	}
	if ClassifyFlow(-1, -15, 15) != FlowContinuation {
		t.Fatal("SHORT negative")
	}
	if ClassifyFlow(1, 0, 15) != FlowNeutral {
		t.Fatal("neutral")
	}
	if ClassifyFlow(0, 40, 15) != FlowNeutral {
		t.Fatal("no legacy dir")
	}
}

func TestSpecHashImmutable(t *testing.T) {
	raw := []byte(`{"spec_id":"FLOW_EXHAUSTION_V1","threshold":15,"spec_hash":"PENDING"}`)
	h1, err := CanonicalSpecHash(raw)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	m["spec_hash"] = h1
	b, _ := json.Marshal(m)
	h2, err := CanonicalSpecHash(b)
	if err != nil || h2 != h1 {
		t.Fatalf("hash must ignore spec_hash field %s %s", h1, h2)
	}
	m["threshold"] = 16.0
	b, _ = json.Marshal(m)
	h3, _ := CanonicalSpecHash(b)
	if h3 == h1 {
		t.Fatal("mutating spec must change hash")
	}
}

func TestDatasetOverlapAndHoldout(t *testing.T) {
	disc := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	discEnd := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	from := time.Date(2026, 3, 17, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 14, 23, 59, 59, 0, time.UTC)
	if OverlapsDiscovery(from, to, disc, discEnd) {
		t.Fatal("holdout must not overlap discovery")
	}
	if !HoldoutValid(from, to, disc) {
		t.Fatal("holdout valid")
	}
	if HoldoutValid(from, disc.Add(time.Hour), disc) {
		t.Fatal("must end before discovery")
	}
	if !OverlapsDiscovery(disc, discEnd, disc, discEnd) {
		t.Fatal("discovery overlaps itself")
	}
}

func TestMinimumNAndSubperiodDecision(t *testing.T) {
	base := BucketStats{N: 100, Mean: 0.0001, MAE: -0.001}
	small := BucketStats{N: 49, Mean: 0.002, MAE: -0.001}
	if DecideExhaustion(small, base, []int{1, 1, 1}) != "INCONCLUSIVE" {
		t.Fatal("n<50")
	}
	good := BucketStats{N: 50, Mean: 0.001, MAE: -0.001}
	if DecideExhaustion(good, base, []int{1, 1, -1}) != "SUPPORTED" {
		t.Fatal("2/3 sign")
	}
	if DecideExhaustion(good, base, []int{-1, -1, 1}) != "INCONCLUSIVE" {
		t.Fatal("unstable")
	}
	worse := BucketStats{N: 50, Mean: 0.00005, MAE: -0.001}
	if DecideExhaustion(worse, base, []int{-1, -1, -1}) != "NOT_SUPPORTED" {
		t.Fatal("effect<=0")
	}
	crash := BucketStats{N: 50, Mean: 0.002, MAE: -0.003}
	if DecideExhaustion(crash, base, []int{1, 1, 1}) != "NOT_SUPPORTED" {
		t.Fatal("mae catastrophe")
	}
}

func TestRepoSpecFrozen(t *testing.T) {
	p := filepath.Join("..", "..", "research", "specs", "FLOW_EXHAUSTION_V1.json")
	h, raw, err := LoadSpec(p)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	stored, _ := m["spec_hash"].(string)
	if stored == "" || stored == "PENDING" || stored != h {
		t.Fatalf("stored hash %s != canonical %s", stored, h)
	}
	if m["spec_id"] != SpecIDV1 {
		t.Fatal(m["spec_id"])
	}
}

func TestProspectiveImmutable(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "signals.jsonl")
	r := ProspectiveRecord{
		RecordedAt: time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC),
		SignalTime: time.Date(2026, 9, 13, 0, 1, 0, 0, time.UTC),
		SpecID:     SpecIDV1, SpecHash: "abc", LegacyDirection: 1,
		OriginalPressure: -20, DirPressure: -20, FlowInterp: FlowExhaustion,
	}
	if err := AppendProspective(p, r); err != nil {
		t.Fatal(err)
	}
	if err := AppendProspective(p, r); err == nil {
		t.Fatal("duplicate signal time must refuse rewrite")
	}
	raw, _ := os.ReadFile(p)
	if len(raw) == 0 {
		t.Fatal("empty")
	}
}
