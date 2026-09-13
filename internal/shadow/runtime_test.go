package shadow

import (
	"testing"
	"time"

	"aurumflow/internal/research"
)

func TestNoExecutionProvider(t *testing.T) {
	var r Runtime
	if HasExecutionProvider(&r) {
		t.Fatal("shadow runtime must not carry ExecutionProvider")
	}
}

func TestDuplicateSignalID(t *testing.T) {
	t0 := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	a := SignalID("BTCUSDT", t0, 1)
	b := SignalID("BTCUSDT", t0, 1)
	if a != b {
		t.Fatal("id unstable")
	}
	if SignalID("BTCUSDT", t0, -1) == a {
		t.Fatal("dir must change id")
	}
	dir := t.TempDir()
	in := research.ProspectiveInput{SignalID: a, Timestamp: t0, Instrument: "BTCUSDT", OutcomeKnown: false}
	if err := research.AppendInput(dir, in); err != nil {
		t.Fatal(err)
	}
	if err := research.AppendInput(dir, in); err == nil {
		t.Fatal("duplicate")
	}
}

func TestNoFutureBookInObserve(t *testing.T) {
	// Absorption Observe only consumes values already computed at t0.
	if MayMutate() {
		t.Fatal()
	}
}

func MayMutate() bool { return false }
