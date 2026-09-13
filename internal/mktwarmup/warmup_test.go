package mktwarmup

import (
	"testing"
	"time"

	"aurumflow/internal/livesurface"
)

func TestWarmupClosedIsNotError(t *testing.T) {
	h := ClosedUnavailable("GOLD")
	if h.Status != StatusUnavailable {
		t.Fatal(h)
	}
}

func TestAssessReadyVsWarming(t *testing.T) {
	now := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	var m5, h1, h4 []livesurface.Candle
	for i := 0; i < 60; i++ {
		m5 = append(m5, livesurface.Candle{Time: now.Add(-time.Duration(60-i) * 5 * time.Minute), Close: 2000})
	}
	for i := 0; i < 24; i++ {
		h1 = append(h1, livesurface.Candle{Time: now.Add(-time.Duration(24-i) * time.Hour), Close: 2000})
	}
	for i := 0; i < 12; i++ {
		h4 = append(h4, livesurface.Candle{Time: now.Add(-time.Duration(12-i) * 4 * time.Hour), Close: 2000})
	}
	h := FromBars("GOLD", m5, h1, h4, now)
	if h.Status != StatusReady || h.M5 != 60 {
		t.Fatal(h)
	}
	if FromBars("US100", m5[:3], nil, nil, now).Status != StatusWarming {
		t.Fatal("short history")
	}
}
