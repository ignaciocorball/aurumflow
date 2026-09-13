package candles

import (
	"testing"
	"time"
)

func TestSynthesizeOHLC(t *testing.T) {
	t0 := time.Date(2026, 8, 1, 0, 0, 10, 0, time.UTC)
	cs := Synthesize([]Trade{
		{t0, 100, 1},
		{t0.Add(20 * time.Second), 102, 1},
		{t0.Add(40 * time.Second), 99, 2},
		{t0.Add(70 * time.Second), 101, 1},
	}, time.Minute)
	if len(cs) != 2 {
		t.Fatalf("n=%d", len(cs))
	}
	if cs[0].Open != 100 || cs[0].High != 102 || cs[0].Low != 99 || cs[0].Close != 99 || cs[0].Volume != 4 {
		t.Fatalf("%+v", cs[0])
	}
	if cs[1].Open != 101 || cs[1].Time.Minute() != 1 {
		t.Fatalf("%+v", cs[1])
	}
}
