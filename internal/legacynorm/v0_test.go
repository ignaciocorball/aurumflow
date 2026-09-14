package legacynorm

import (
	"testing"
	"time"

	"aurumflow/pkg/models"
)

func bars(n int, px, rng float64) []models.Candle {
	var out []models.Candle
	t0 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < n; i++ {
		w := rng
		switch {
		case i < 20:
			w = rng * 0.3
		case i < 40:
			w = rng * 2
		}
		c := px
		out = append(out, models.Candle{
			Time: t0.Add(time.Duration(i) * 15 * time.Minute),
			Open: c, High: c + w, Low: c - w, Close: c,
		})
	}
	return out
}

func TestVolSuitablePastOnly(t *testing.T) {
	cs := bars(80, 2000, 10)
	if !VolSuitable(cs[:70]) {
		t.Fatal("mid-range own ATR% should pass")
	}
	if _, ok := ATRPctRank(cs[:20]); ok {
		t.Fatal("short history")
	}
}

func TestNormalizedInvariantAcrossScale(t *testing.T) {
	a := bars(80, 100, 1)
	b := bars(80, 10000, 100)
	ra, oka := ATRPctRank(a)
	rb, okb := ATRPctRank(b)
	if !oka || !okb {
		t.Fatal("rank")
	}
	if ra < 0.1 || rb < 0.1 {
		t.Fatalf("scale invariance ra=%v rb=%v", ra, rb)
	}
}

func TestSpecFrozen(t *testing.T) {
	if Spec != "LEGACY_NORMALIZED_V0" || SpecHash != "sha256:f4c22591fbb0086c3a0cde0d49f1f54f5b5d97b30bfc1c811c563bd59d78387f" {
		t.Fatal(Spec, SpecHash)
	}
}

func TestGOLDCompat(t *testing.T) {
	if Compat("GOLD", 12) != "MECHANICALLY_COMPATIBLE" {
		t.Fatal(Compat("GOLD", 12))
	}
	if Compat("J225", 400) != "SESSION_INCOMPATIBLE" {
		t.Fatal(Compat("J225", 400))
	}
	if Compat("SILVER", 0.4) != "SCALE_INCOMPATIBLE" {
		t.Fatal(Compat("SILVER", 0.4))
	}
}

func TestNoFutureBarInRank(t *testing.T) {
	cs := bars(80, 100, 1)
	cs[79].High = 1e9
	r1, _ := ATRPctRank(cs[:70])
	r2, _ := ATRPctRank(cs[:70])
	if r1 != r2 {
		t.Fatal("deterministic")
	}
}
