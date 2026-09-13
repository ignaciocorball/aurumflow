package xasset

import "math"

const Freshness = "PROXY_CROSS_ASSET"

type Frame struct {
	Symbol string
	Close  []float64
}

func Returns(close []float64) []float64 {
	if len(close) < 2 {
		return nil
	}
	out := make([]float64, len(close)-1)
	for i := 1; i < len(close); i++ {
		if close[i-1] == 0 {
			continue
		}
		out[i-1] = (close[i] - close[i-1]) / close[i-1]
	}
	return out
}

func RelativeStrength(a, b []float64) float64 {
	ra, rb := last(Returns(a)), last(Returns(b))
	return ra - rb
}

func RollingCorr(a, b []float64, win int) float64 {
	if win < 3 || len(a) < win || len(b) < win {
		return 0
	}
	aa := a[len(a)-win:]
	bb := b[len(b)-win:]
	ma, mb := mean(aa), mean(bb)
	var num, da, db float64
	for i := range aa {
		x, y := aa[i]-ma, bb[i]-mb
		num += x * y
		da += x * x
		db += y * y
	}
	den := math.Sqrt(da * db)
	if den == 0 {
		return 0
	}
	return num / den
}

func RollingBeta(y, x []float64, win int) float64 {
	if win < 3 || len(y) < win || len(x) < win {
		return 0
	}
	yy := y[len(y)-win:]
	xx := x[len(x)-win:]
	mx, my := mean(xx), mean(yy)
	var cov, vx float64
	for i := range xx {
		dx := xx[i] - mx
		cov += dx * (yy[i] - my)
		vx += dx * dx
	}
	if vx == 0 {
		return 0
	}
	return cov / vx
}

func Divergence(a, b []float64) bool {
	if len(a) < 2 || len(b) < 2 {
		return false
	}
	return (a[len(a)-1]-a[len(a)-2])*(b[len(b)-1]-b[len(b)-2]) < 0
}

func Breadth(dirs []int) (leaders, laggards int, agreement float64) {
	if len(dirs) == 0 {
		return 0, 0, 0
	}
	pos, neg := 0, 0
	for _, d := range dirs {
		if d > 0 {
			pos++
		} else if d < 0 {
			neg++
		}
	}
	leaders, laggards = pos, neg
	if pos+neg == 0 {
		return leaders, laggards, 0
	}
	if pos >= neg {
		agreement = float64(pos) / float64(pos+neg)
	} else {
		agreement = -float64(neg) / float64(pos+neg)
	}
	return leaders, laggards, agreement
}

func last(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	return xs[len(xs)-1]
}

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := 0.0
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}
