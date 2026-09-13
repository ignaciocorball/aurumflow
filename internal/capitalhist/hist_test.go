package capitalhist

import (
	"testing"
	"time"

	"aurumflow/pkg/models"
)

func TestDedupGaps(t *testing.T) {
	t0 := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	cs := DedupSort([]models.Candle{
		{Time: t0.Add(time.Hour), Close: 2},
		{Time: t0, Close: 1},
		{Time: t0, Close: 1},
	})
	if len(cs) != 2 || !cs[0].Time.Equal(t0) {
		t.Fatalf("%+v", cs)
	}
	g := CountGaps([]models.Candle{{Time: t0}, {Time: t0.Add(3 * time.Hour)}}, time.Hour)
	if g == 0 {
		t.Fatal("gap")
	}
}
