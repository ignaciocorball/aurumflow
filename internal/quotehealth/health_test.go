package quotehealth

import (
	"testing"
	"time"

	"aurumflow/internal/livesurface"
	"aurumflow/internal/worlddomain"
)

func TestFreshnessStates(t *testing.T) {
	now := time.Now().UTC()
	live := livesurface.NewQuote("GOLD", "GOLD", 1, 2, "TRADEABLE", now, now, time.Minute)
	if Of(live, 12, time.Minute, 0).Status != worlddomain.HealthHealthy {
		t.Fatal("live")
	}
	stale := livesurface.NewQuote("GOLD", "GOLD", 1, 2, "TRADEABLE", now.Add(-10*time.Minute), now, time.Minute)
	if Of(stale, 0, time.Minute, 1).Status != worlddomain.HealthStale {
		t.Fatal("stale")
	}
	closed := livesurface.NewQuote("US100", "US100", 1, 2, "CLOSED", now, now, time.Minute)
	if TapeState(closed, "READY") != "CLOSED" {
		t.Fatal(TapeState(closed, "READY"))
	}
	if TapeState(closed, "WARMING") != "WARMING" {
		t.Fatal("warmup while closed")
	}
}
