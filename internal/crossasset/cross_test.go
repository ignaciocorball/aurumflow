package crossasset

import (
	"testing"

	"aurumflow/internal/xasset"
)

func TestFramesNoCausalityClaim(t *testing.T) {
	us := []float64{100, 101, 102, 103}
	gold := xasset.Frame{Symbol: "GOLD", Close: []float64{2000, 2010, 2020, 2040}}
	snaps := FromFrames([]xasset.Frame{gold}, us)
	if len(snaps) != 1 || !snaps[0].Present {
		t.Fatal(snaps)
	}
	l, g, a := MegaCapBreadth([]int{1, 1, 1, -1, 1})
	if l+g != 5 || a <= 0 {
		t.Fatal(l, g, a)
	}
}
