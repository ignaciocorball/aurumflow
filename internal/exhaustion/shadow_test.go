package exhaustion

import (
	"testing"
	"time"
)

func TestShadowCannotMutateBroker(t *testing.T) {
	e := NewEngine("BTCUSDT")
	if e.Mode != ModeShadow {
		t.Fatal(e.Mode)
	}
	s := e.Observe(time.Now().UTC(), 1, 6, -20, 1)
	if s.Mode != ModeShadow || s.Classification != ClassExhaustion {
		t.Fatal(s)
	}
	if MayMutateBroker() {
		t.Fatal("broker mutation")
	}
}

func TestMatchNoFuture(t *testing.T) {
	t0 := time.Date(2026, 5, 1, 15, 0, 0, 0, time.UTC)
	rows := []MatchRow{
		{Time: t0, Dir: 1, Hour: 15, Score: 6, Vol: 0.01, Class: ClassExhaustion, Idx: 0},
		{Time: t0.Add(time.Hour), Dir: 1, Hour: 16, Score: 6, Vol: 0.01, Class: ClassNeutral, Idx: 1},
		{Time: t0.Add(-time.Hour), Dir: 1, Hour: 15, Score: 6, Vol: 0.01, Class: ClassNeutral, Idx: 2},
	}
	pairs := MatchControls(rows, []float64{0.01, 0.01, 0.01})
	if len(pairs) != 1 || pairs[0][1] != 2 {
		t.Fatalf("%v", pairs)
	}
}
