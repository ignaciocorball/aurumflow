package worldstate

import (
	"testing"
	"time"

	"aurumflow/internal/livesurface"
	"aurumflow/internal/microcap"
	"aurumflow/internal/worlddomain"
)

func TestBTCMicroWhenRuntimeHealthy(t *testing.T) {
	now := time.Date(2026, 9, 13, 18, 0, 0, 0, time.UTC)
	ws := At(now, Input{Micro: microcap.BTCFromRuntime(true, true, true, true, true, "HEALTHY")})
	if !ws.Markets["BTC"].MicroAvailable {
		t.Fatal("BTC micro should advertise runtime capability")
	}
}

func TestClosedLiveFrameDoesNotRankMomentum(t *testing.T) {
	now := time.Date(2026, 9, 13, 18, 0, 0, 0, time.UTC)
	q := livesurface.NewQuote("US100", "US100", 20000, 20002, "CLOSED", now, now, time.Minute)
	frame := livesurface.Assemble(now, map[string]livesurface.Quote{"US100": q}, map[string][]livesurface.Candle{
		"US100": {{Time: now.Add(-time.Hour), Close: 19800}, {Time: now, Close: 20000}},
	})
	ws := At(now, Input{})
	ws = ApplyLiveFrame(ws, frame)
	if ws.Markets["US100"].PriceTrend == "UP" || ws.Markets["US100"].PriceTrend == "STRONG_UP" {
		t.Fatal(ws.Markets["US100"].PriceTrend)
	}
	if ws.Markets["US100"].DataQuality != worlddomain.HealthDegraded {
		t.Fatal(ws.Markets["US100"].DataQuality)
	}
}
