package research

import "math"

func RelativeStrength(a, b []float64) []float64 {
	n := min(len(a), len(b))
	out := make([]float64, n)
	for i := 1; i < n; i++ {
		if a[i-1] == 0 || b[i-1] == 0 {
			continue
		}
		ra := (a[i] - a[i-1]) / a[i-1]
		rb := (b[i] - b[i-1]) / b[i-1]
		out[i] = ra - rb
	}
	return out
}

func RollingCorr(a, b []float64, win int) float64 {
	n := min(len(a), len(b))
	if win <= 2 || n < win {
		return 0
	}
	a = a[n-win:]
	b = b[n-win:]
	ma, mb := mean(a), mean(b)
	num, da, db := 0.0, 0.0, 0.0
	for i := range a {
		x, y := a[i]-ma, b[i]-mb
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

func mean(xs []float64) float64 {
	s := 0.0
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
