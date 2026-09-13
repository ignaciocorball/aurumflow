package research

import (
	"testing"
	"time"

	"aurumflow/internal/exhaustion"
	"aurumflow/internal/radar"
	"aurumflow/pkg/models"
)

func TestMechanismPastOnlyAndGroups(t *testing.T) {
	t0 := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	var m1 []models.Candle
	var snaps []radar.PressureSnapshot
	for i := 0; i < 40; i++ {
		tt := t0.Add(time.Duration(i) * time.Minute)
		m1 = append(m1, models.Candle{Time: tt, Open: 100, High: 101, Low: 99, Close: 100, Volume: 10})
		snaps = append(snaps, radar.PressureSnapshot{Timestamp: tt, PressureScore: -20, AggBuy: 1, AggSell: 8, CVD: -float64(i), TradeVel: 2})
	}
	fused := []SignalRow{{Time: t0.Add(30 * time.Minute), Direction: 1, Score: 6, OriginalPressure: -20, FlowInterp: exhaustion.ClassExhaustion}}
	rows := FeaturesFromTape(fused, m1, snaps)
	if len(rows) != 1 {
		t.Fatalf("rows=%d", len(rows))
	}
	if rows[0].Snap.Classification != exhaustion.ClassExhaustion {
		t.Fatal(rows[0].Snap.Classification)
	}
	if rows[0].Snap.Features.At.After(fused[0].Time) {
		t.Fatal("lookahead")
	}
}
