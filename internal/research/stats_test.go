package research

import "testing"

func TestSummarizeAndBootstrap(t *testing.T) {
	s := Summarize([]float64{0.01, -0.01, 0.02}, []float64{0.02, 0.01, 0.03}, []float64{-0.01, -0.02, -0.01})
	if s.N != 3 || s.Hit < 0.6 || s.Median == 0 {
		t.Fatalf("%+v", s)
	}
	lo, hi := BootstrapMeanCI([]float64{1, 1, 1, 1}, 1)
	if lo > 1.01 || hi < 0.99 {
		t.Fatalf("%v %v", lo, hi)
	}
}
