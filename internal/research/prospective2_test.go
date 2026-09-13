package research

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProspectiveInputOutcomeSeparation(t *testing.T) {
	dir := t.TempDir()
	in := ProspectiveInput{
		SignalID: "sig-1", RecordedAt: time.Now().UTC(), Timestamp: time.Now().UTC(),
		Instrument: "BTCUSDT", LegacyDirection: 1, PressureScore: -20,
		V1Classification: "FLOW_EXHAUSTION_CONFIRM", SpecHash: "abc", OutcomeKnown: false,
	}
	if err := AppendInput(dir, in); err != nil {
		t.Fatal(err)
	}
	if err := AppendInput(dir, in); err == nil {
		t.Fatal("duplicate")
	}
	bad := in
	bad.SignalID = "sig-2"
	bad.OutcomeKnown = true
	if err := AppendInput(dir, bad); err == nil {
		t.Fatal("outcome_known on input")
	}
	if err := AppendOutcome(dir, ProspectiveOutcome{SignalID: "missing"}); err == nil {
		t.Fatal("unknown id")
	}
	if err := AppendOutcome(dir, ProspectiveOutcome{SignalID: "sig-1", ObservedAt: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	if err := AppendOutcome(dir, ProspectiveOutcome{SignalID: "sig-1"}); err == nil {
		t.Fatal("dup outcome")
	}
	_ = filepath.Separator
}

func TestLabelDueDoesNotMutateInput(t *testing.T) {
	dir := t.TempDir()
	t0 := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	in := ProspectiveInput{
		SignalID: "sig-lab", RecordedAt: t0, Timestamp: t0,
		Instrument: "BTCUSDT", LegacyDirection: 1, PressureScore: -20,
		V1Classification: "FLOW_EXHAUSTION_CONFIRM", OutcomeKnown: false,
	}
	if err := AppendInput(dir, in); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(InputPath(dir))
	prices := []PricePoint{{T: t0, P: 100}, {T: t0.Add(time.Minute), P: 100.2}, {T: t0.Add(15 * time.Minute), P: 100.5}, {T: t0.Add(time.Hour), P: 101}}
	n, err := LabelDue(dir, prices, t0.Add(2*time.Hour))
	if err != nil || n != 1 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	after, _ := os.ReadFile(InputPath(dir))
	if string(before) != string(after) {
		t.Fatal("input mutated")
	}
	if !outcomeExists(OutcomePath(dir), "sig-lab") {
		t.Fatal("outcome missing")
	}
}
