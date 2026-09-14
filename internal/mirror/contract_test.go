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

func TestProtectionMismatchHalts(t *testing.T) {
	if !ProtectionOK("US100", "US100", "BUY", "BUY", 0.01, 0.01, 100, 100, 110, 110, 0.05) {
		t.Fatal("match")
	}
	if ProtectionOK("US100", "GOLD", "BUY", "BUY", 0.01, 0.01, 100, 100, 110, 110, 0.05) {
		t.Fatal("epic")
	}
	if ProtectionOK("US100", "US100", "BUY", "BUY", 0.01, 0.01, 100, 90, 110, 110, 0.05) {
		t.Fatal("sl")
	}
}

func TestEntrySlippage(t *testing.T) {
	if d := EntrySlippage("BUY", 10, 10.2, 10.3); d < 0.099 || d > 0.101 {
		t.Fatalf("buy %v", d)
	}
	if d := EntrySlippage("SELL", 10, 10.2, 9.9); d < 0.099 || d > 0.101 {
		t.Fatalf("sell %v", d)
	}
}
