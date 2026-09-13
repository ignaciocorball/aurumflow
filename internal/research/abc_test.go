package research

import (
	"testing"
	"time"

	"aurumflow/internal/radar"
	"aurumflow/pkg/models"
)

func TestAlignAndFilter(t *testing.T) {
	t0 := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	legacy := []SignalRow{{Time: t0.Add(time.Hour), Direction: 1, Source: "legacy"}}
	rp := []RadarPoint{{Time: t0.Add(30 * time.Minute), Pressure: 40, Direction: 1, State: radar.StateExpansion, Confidence: 70}}
	got := AlignFusion(legacy, rp, 15)
	if got[0].Class != AlignedLong {
		t.Fatal(got[0].Class)
	}
	px := []PricePoint{{t0, 100}, {t0.Add(2 * time.Hour), 101}}
	all, filt := FilterValue(got, px, time.Hour)
	if all.N == 0 && filt.N == 0 {
		t.Fatal("expected some labels if complete")
	}
	_ = models.Candle{}
}
