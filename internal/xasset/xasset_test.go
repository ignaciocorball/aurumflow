package xasset

import "testing"

func TestCrossAssetBasics(t *testing.T) {
	gold := []float64{100, 101, 102, 101, 103}
	silver := []float64{20, 20.2, 20.1, 20.0, 20.4}
	if len(Returns(gold)) != 4 {
		t.Fatal("returns")
	}
	if RelativeStrength(gold, silver) == 0 && gold[len(gold)-1] == silver[len(silver)-1] {
		t.Fatal("rs")
	}
	if RollingCorr(gold, gold, 4) < 0.99 {
		t.Fatal("corr self")
	}
	_ = RollingBeta(gold, silver, 4)
	if !Divergence([]float64{1, 2}, []float64{2, 1}) {
		t.Fatal("div")
	}
	l, g, agr := Breadth([]int{1, 1, -1, 0})
	if l != 2 || g != 1 || agr <= 0 {
		t.Fatalf("%d %d %v", l, g, agr)
	}
}
