package research

import "math"

type BucketStats struct {
	Name   string  `json:"name"`
	N      int     `json:"n"`
	Hit    float64 `json:"hit_rate"`
	Mean   float64 `json:"mean"`
	MeanLo float64 `json:"mean_ci_lo,omitempty"`
	MeanHi float64 `json:"mean_ci_hi,omitempty"`
	Median float64 `json:"median"`
	MFE    float64 `json:"mfe"`
	MAE    float64 `json:"mae"`
	Ratio  float64 `json:"mfe_mae"`
}

func Summarize(dirRets, mfes, maes []float64) BucketStats {
	s := BucketStats{N: len(dirRets)}
	if s.N == 0 {
		return s
	}
	hits := 0
	sum := 0.0
	mfe, mae := 0.0, 0.0
	cp := append([]float64(nil), dirRets...)
	for i, r := range dirRets {
		sum += r
		if r > 0 {
			hits++
		}
		if i < len(mfes) {
			mfe += mfes[i]
		}
		if i < len(maes) {
			mae += maes[i]
		}
	}
	s.Hit = float64(hits) / float64(s.N)
	s.Mean = sum / float64(s.N)
	s.Median = median(cp)
	s.MFE = mfe / float64(s.N)
	s.MAE = mae / float64(s.N)
	s.Ratio = MFEMAERatio(s.MFE, s.MAE)
	s.MeanLo, s.MeanHi = BootstrapMeanCI(dirRets, 42)
	return s
}

func median(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	for i := 1; i < len(xs); i++ {
		j := i
		for j > 0 && xs[j] < xs[j-1] {
			xs[j], xs[j-1] = xs[j-1], xs[j]
			j--
		}
	}
	if len(xs)%2 == 1 {
		return xs[len(xs)/2]
	}
	return (xs[len(xs)/2-1] + xs[len(xs)/2]) / 2
}

func BootstrapMeanCI(xs []float64, seed int64) (lo, hi float64) {
	if len(xs) == 0 {
		return 0, 0
	}
	const n = 200
	means := make([]float64, n)
	x := seed
	for i := 0; i < n; i++ {
		sum := 0.0
		for j := 0; j < len(xs); j++ {
			x = x*1664525 + 1013904223
			idx := int(uint32(x) % uint32(len(xs)))
			sum += xs[idx]
		}
		means[i] = sum / float64(len(xs))
	}
	for i := 1; i < n; i++ {
		j := i
		for j > 0 && means[j] < means[j-1] {
			means[j], means[j-1] = means[j-1], means[j]
			j--
		}
	}
	return means[int(0.025*n)], means[int(0.975*n)]
}

func Abs(v float64) float64 { return math.Abs(v) }
