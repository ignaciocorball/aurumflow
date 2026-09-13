package research

import (
	"testing"
	"time"
)

func TestForwardLabelsMFEMAE(t *testing.T) {
	t0 := time.Unix(0, 0).UTC()
	px := []PricePoint{
		{t0, 100},
		{t0.Add(10 * time.Second), 101},
		{t0.Add(20 * time.Second), 99},
		{t0.Add(30 * time.Second), 102},
	}
	ls := Labels(px, t0, 100, 1, []time.Duration{30 * time.Second})
	if len(ls) != 1 || ls[0].ForwardRet < 0.01 || ls[0].MFE < 0.01 || ls[0].MAE >= 0 {
		t.Fatalf("%+v", ls[0])
	}
	if MFEMAERatio(0.02, -0.01) != 2 {
		t.Fatal("ratio")
	}
}
