package radar

import (
	"testing"
	"time"
)

func TestComposeExplainsScore(t *testing.T) {
	s := Compose(PressureSnapshot{
		AggressiveFlowScore: 18, AbsorptionScore: 24, BookImbalanceScore: 12,
		LiquidityScore: 10, PersistenceScore: 7, VolatilityContext: 3,
		BookSynced: true, Confidence: 80,
	})
	if s.PressureScore != 74 {
		t.Fatalf("score=%v", s.PressureScore)
	}
	if len(s.Evidence) != 6 {
		t.Fatal("evidence")
	}
	if s.Direction != 1 {
		t.Fatal("dir")
	}
}

func TestUnsyncedBookNoTrade(t *testing.T) {
	s := Compose(PressureSnapshot{AggressiveFlowScore: 90, BookSynced: false, Confidence: 80})
	if s.State != StateNoTrade || s.Confidence >= 80 {
		t.Fatalf("%+v", s)
	}
}

func TestHysteresis(t *testing.T) {
	h := Hysteresis{MinHold: time.Second}
	now := time.Unix(0, 0).UTC()
	if !h.Allow(StateExpansion, now) {
		t.Fatal("first")
	}
	if h.Allow(StateNoTrade, now.Add(100*time.Millisecond)) {
		t.Fatal("too soon")
	}
	if !h.Allow(StateNoTrade, now.Add(2*time.Second)) {
		t.Fatal("after hold")
	}
}
