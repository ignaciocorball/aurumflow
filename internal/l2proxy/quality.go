package l2proxy

import "math"

const (
	Good      = "GOOD"
	Degraded  = "DEGRADED"
	Unusable  = "UNUSABLE"
)

// Input uses operational thresholds only. Not fit to trading outcomes.
type Input struct {
	BookSynced       bool
	BookAgeMs        int64
	ProviderOK       bool
	ReceiveLatencyP95 float64
	AbsBasisZ        float64
	ClockSkewMs      int64
}

type Result struct {
	Quality string   `json:"quality"`
	Reasons []string `json:"reasons,omitempty"`
}

func Classify(in Input) Result {
	var r Result
	if !in.ProviderOK || !in.BookSynced {
		r.Quality = Unusable
		if !in.ProviderOK {
			r.Reasons = append(r.Reasons, "provider_unhealthy")
		}
		if !in.BookSynced {
			r.Reasons = append(r.Reasons, "book_unsynced")
		}
		return r
	}
	if in.BookAgeMs > 5000 || in.ReceiveLatencyP95 > 2000 || in.AbsBasisZ > 6 || abs64(in.ClockSkewMs) > 3000 {
		r.Quality = Unusable
		if in.BookAgeMs > 5000 {
			r.Reasons = append(r.Reasons, "book_stale")
		}
		if in.ReceiveLatencyP95 > 2000 {
			r.Reasons = append(r.Reasons, "latency")
		}
		if in.AbsBasisZ > 6 {
			r.Reasons = append(r.Reasons, "basis_dislocated")
		}
		if abs64(in.ClockSkewMs) > 3000 {
			r.Reasons = append(r.Reasons, "clock_skew")
		}
		return r
	}
	if in.BookAgeMs > 1500 || in.ReceiveLatencyP95 > 500 || in.AbsBasisZ > 3 || abs64(in.ClockSkewMs) > 1000 {
		r.Quality = Degraded
		if in.BookAgeMs > 1500 {
			r.Reasons = append(r.Reasons, "book_aging")
		}
		if in.ReceiveLatencyP95 > 500 {
			r.Reasons = append(r.Reasons, "latency")
		}
		if in.AbsBasisZ > 3 {
			r.Reasons = append(r.Reasons, "basis")
		}
		if abs64(in.ClockSkewMs) > 1000 {
			r.Reasons = append(r.Reasons, "clock")
		}
		return r
	}
	r.Quality = Good
	return r
}

func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

type Basis struct {
	hist []float64
	cap  int
}

func NewBasis(cap int) *Basis {
	if cap <= 0 {
		cap = 300
	}
	return &Basis{cap: cap}
}

func (b *Basis) Observe(midA, midB float64) (basis, z float64) {
	if midA <= 0 || midB <= 0 {
		return 0, 0
	}
	basis = midB - midA
	z = robustZ(b.hist, basis)
	b.hist = append(b.hist, basis)
	if len(b.hist) > b.cap {
		b.hist = b.hist[len(b.hist)-b.cap:]
	}
	return basis, z
}

func robustZ(xs []float64, v float64) float64 {
	if len(xs) < 8 {
		return 0
	}
	cp := append([]float64(nil), xs...)
	for i := 1; i < len(cp); i++ {
		j := i
		for j > 0 && cp[j] < cp[j-1] {
			cp[j], cp[j-1] = cp[j-1], cp[j]
			j--
		}
	}
	med := cp[len(cp)/2]
	dev := make([]float64, len(cp))
	for i, x := range cp {
		dev[i] = math.Abs(x - med)
	}
	for i := 1; i < len(dev); i++ {
		j := i
		for j > 0 && dev[j] < dev[j-1] {
			dev[j], dev[j-1] = dev[j-1], dev[j]
			j--
		}
	}
	mad := dev[len(dev)/2]
	if mad <= 1e-12 {
		if v == med {
			return 0
		}
		if v > med {
			return 10
		}
		return -10
	}
	return (v - med) / (1.4826 * mad)
}
