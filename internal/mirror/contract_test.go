package mirror

import (
	"testing"

	"aurumflow/internal/research"
)

func TestRejectsAttentionSalienceAndFake(t *testing.T) {
	sig := research.SignalRow{Direction: 1}
	if err := AllowOrder(Intent{Source: SourceAttention, Signal: sig}); err != ErrResearchOnly {
		t.Fatal(err)
	}
	if err := AllowOrder(Intent{Source: SourceSalience, Signal: sig}); err != ErrResearchOnly {
		t.Fatal(err)
	}
	if err := AllowOrder(Intent{Source: SourceNormalized, Signal: sig, Fake: true}); err != ErrFake {
		t.Fatal(err)
	}
	if err := AllowOrder(Intent{Source: SourceNormalized, Signal: research.SignalRow{}}); err != ErrNoSignal {
		t.Fatal(err)
	}
}

func TestAllowsFrozenStrategyOnly(t *testing.T) {
	if err := AllowOrder(Intent{Source: SourceNormalized, Signal: research.SignalRow{Direction: -1}}); err != nil {
		t.Fatal(err)
	}
	if err := AllowOrder(Intent{Source: "MANUAL", Signal: research.SignalRow{Direction: 1}}); err != ErrSource {
		t.Fatal(err)
	}
}

func TestOneSignalOneOrder(t *testing.T) {
	seen := map[string]bool{}
	if err := OneSignalOneOrder(seen, "US100-1-BUY"); err != nil {
		t.Fatal(err)
	}
	seen["US100-1-BUY"] = true
	if err := OneSignalOneOrder(seen, "US100-1-BUY"); err != ErrDuplicate {
		t.Fatal(err)
	}
}

func TestReconcileHalt(t *testing.T) {
	if Reconcile(0, 0, 0) != nil {
		t.Fatal("clean")
	}
	if !HaltOnMismatch(1, 0, 0) || !HaltOnMismatch(0, 1, 0) || !HaltOnMismatch(0, 0, 1) {
		t.Fatal("mismatch")
	}
}
