package researchopp

import (
	"testing"
	"time"
)

func TestLabelClosedGapUnavailable(t *testing.T) {
	t0 := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	if LabelQuality(t0, time.Hour, nil, true) != OutcomeUnavail {
		t.Fatal("gap")
	}
	path := []PricePoint{{T: t0.Add(time.Hour), P: 10}}
	if LabelQuality(t0, time.Hour, path, false) != OutcomeValid && LabelQuality(t0, time.Hour, path, false) != OutcomePartial {
		q := LabelQuality(t0, time.Hour, path, false)
		if q == OutcomeUnavail {
			t.Fatal(q)
		}
	}
}

func TestForwardReturnPastOnly(t *testing.T) {
	t0 := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	path := []PricePoint{{T: t0.Add(15 * time.Minute), P: 110}}
	r, ok := ForwardReturn(t0, 15*time.Minute, 100, path)
	if !ok || r < 0.09 {
		t.Fatal(r, ok)
	}
}
