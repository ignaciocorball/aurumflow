package salience

import (
	"math"
	"sort"
)

const Spec = "MARKET_SALIENCE_V0"
const SpecHash = "sha256:f6747a40b0fc2d758b7a746e8663121e45e607f329222b6b8b96d68b5efede38"

type Bar struct {
	Close, High, Low float64
}

type Input struct {
	Market     string
	Bars       []Bar // oldest→newest, all bar_end <= t0
	DQPenalty  float64
	CrossAbsRet []float64 // |return| of TRADEABLE peers at t0 (optional)
}

type Result struct {
	Market     string
	Score      float64
	Label      string
	Components map[string]float64
	Reason     string
}

func Score(in Input) Result {
	r := Result{Market: in.Market, Label: "MARKET SALIENCE", Components: map[string]float64{}}
	if len(in.Bars) < 8 {
		r.Reason = "INSUFFICIENT_HISTORY"
		return r
	}
	look := 12
	if len(in.Bars) < look+1 {
		look = len(in.Bars) - 1
	}
	absZ := absRetZ(in.Bars, look)
	rvX := rvExpand(in.Bars)
	rgX := rangeExpand(in.Bars)
	xs := xsPct(absRet(in.Bars, look), in.CrossAbsRet)
	rs := rsDisp(absRet(in.Bars, look), in.CrossAbsRet)
	r.Components["abs_ret_z"] = scale(absZ)
	r.Components["rv_expand"] = scale(rvX)
	r.Components["range_expand"] = scale(rgX)
	r.Components["rs_disp"] = scale(rs)
	r.Components["xs_pct"] = clamp01(xs)
	raw := 0.30*r.Components["abs_ret_z"] + 0.20*r.Components["rv_expand"] + 0.15*r.Components["range_expand"] + 0.15*r.Components["rs_disp"] + 0.20*r.Components["xs_pct"]
	pen := in.DQPenalty
	if pen < 0 {
		pen = 0
	}
	if pen > 1 {
		pen = 1
	}
	r.Score = clamp(100*raw*(1-pen), 0, 100)
	r.Reason = "ok"
	return r
}

func DQPenalty(quality string) float64 {
	switch quality {
	case "HEALTHY":
		return 0
	case "STALE":
		return 0.35
	default:
		return 1
	}
}

func absRet(bars []Bar, look int) float64 {
	n := len(bars)
	if n < look+1 || bars[n-1-look].Close <= 0 {
		return 0
	}
	return math.Abs(bars[n-1].Close-bars[n-1-look].Close) / bars[n-1-look].Close
}

func absRetZ(bars []Bar, look int) float64 {
	ret := absRet(bars, look)
	rv := realizedVol(bars, look)
	if rv <= 1e-12 {
		return 0
	}
	return ret / rv
}

func realizedVol(bars []Bar, look int) float64 {
	n := len(bars)
	if n < look+1 {
		return 0
	}
	var ss float64
	c := 0
	for i := n - look; i < n; i++ {
		if bars[i-1].Close <= 0 {
			continue
		}
		r := (bars[i].Close - bars[i-1].Close) / bars[i-1].Close
		ss += r * r
		c++
	}
	if c < 2 {
		return 0
	}
	return math.Sqrt(ss / float64(c))
}

func rvExpand(bars []Bar) float64 {
	cur := realizedVol(bars, 12)
	if cur <= 0 || len(bars) < 20 {
		return 0
	}
	var meds []float64
	for i := 12; i < len(bars)-1 && len(meds) < 24; i++ {
		v := realizedVol(bars[:i+1], 12)
		if v > 0 {
			meds = append(meds, v)
		}
	}
	m := median(meds)
	if m <= 1e-12 {
		return 0
	}
	return cur / m
}

func rangeExpand(bars []Bar) float64 {
	n := len(bars)
	cur := bars[n-1].High - bars[n-1].Low
	if cur <= 0 {
		return 0
	}
	var rs []float64
	from := n - 25
	if from < 0 {
		from = 0
	}
	for i := from; i < n-1; i++ {
		if w := bars[i].High - bars[i].Low; w > 0 {
			rs = append(rs, w)
		}
	}
	m := median(rs)
	if m <= 1e-12 {
		return 0
	}
	return cur / m
}

func xsPct(own float64, peers []float64) float64 {
	if len(peers) == 0 {
		return 0
	}
	below := 0
	for _, p := range peers {
		if p <= own {
			below++
		}
	}
	return float64(below) / float64(len(peers))
}

func rsDisp(own float64, peers []float64) float64 {
	if len(peers) == 0 {
		return 0
	}
	m := median(peers)
	return math.Abs(own - m)
}

func median(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	cp := append([]float64(nil), xs...)
	sort.Float64s(cp)
	return cp[len(cp)/2]
}

func scale(v float64) float64 {
	if v < 0 {
		v = 0
	}
	if v > 3 {
		v = 3
	}
	return v / 3
}

func clamp01(v float64) float64 {
	return clamp(v, 0, 1)
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
