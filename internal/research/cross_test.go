package research

import "testing"

func TestCrossAssetFeatures(t *testing.T) {
	a := []float64{1, 1.1, 1.2, 1.3}
	b := []float64{1, 1.05, 1.1, 1.12}
	rs := RelativeStrength(a, b)
	if len(rs) != 4 {
		t.Fatal(len(rs))
	}
	c := RollingCorr(a, b, 4)
	if c <= 0 {
		t.Fatalf("corr %v", c)
	}
}
