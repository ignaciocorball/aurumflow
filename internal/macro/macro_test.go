package macro

import (
	"testing"
	"time"
)

func TestMacroNoLookahead(t *testing.T) {
	p := New()
	if p.Status == "" {
		t.Fatal("status")
	}
	t0 := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	p.Points = []SeriesPoint{{ID: "DGS10", Period: t0, AvailableAt: t0.Add(24 * time.Hour), Value: 4.1, Freq: "D"}}
	if _, ok := p.Latest("DGS10", t0); ok {
		t.Fatal("lookahead")
	}
	if _, ok := p.Latest("DGS10", t0.Add(24*time.Hour)); !ok {
		t.Fatal("available")
	}
}
