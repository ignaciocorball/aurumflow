package mktval

import (
	"testing"
	"time"

	"aurumflow/internal/chronosplit"
	"aurumflow/internal/costmodel"
	"aurumflow/internal/research"
	"aurumflow/pkg/models"
)

func TestSimulateNoFutureLeak(t *testing.T) {
	t0 := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	var m15 []models.Candle
	for i := 0; i < 80; i++ {
		px := 100.0 + float64(i)*0.1
		m15 = append(m15, models.Candle{Time: t0.Add(time.Duration(i) * 15 * time.Minute), Open: px, High: px + 1, Low: px - 1, Close: px})
	}
	split := chronosplit.Of(m15[0].Time, m15[len(m15)-1].Time)
	sigs := []research.SignalRow{{Time: m15[30].Time, Direction: 1}}
	tr := Simulate(sigs, m15, 0.1, costmodel.Normal, split)
	if len(tr) != 1 {
		t.Fatal(tr)
	}
	if !tr[0].T0.Equal(m15[30].Time) {
		t.Fatal("t0 rewritten")
	}
}

func TestPromoteRequiresHoldout(t *testing.T) {
	if Promote(Stats{N: 5, Expectancy: 1, PF: 2, MaxDD: 1, PositiveThirds: 3}) {
		t.Fatal("small n")
	}
	if !Promote(Stats{N: 25, Expectancy: 0.1, PF: 1.2, MaxDD: 5, PositiveThirds: 2}) {
		t.Fatal("should promote")
	}
}
