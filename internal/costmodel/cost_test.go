package costmodel

import "testing"

func TestStressWorseThanNormal(t *testing.T) {
	z := Apply(100, 101, 0.2, Zero, 1)
	n := Apply(100, 101, 0.2, Normal, 1)
	s := Apply(100, 101, 0.2, Stress, 1)
	if !(z > n && n > s) {
		t.Fatalf("z=%v n=%v s=%v", z, n, s)
	}
}

func TestShortSigns(t *testing.T) {
	if Apply(100, 99, 0, Zero, -1) <= 0 {
		t.Fatal("short profit")
	}
}
