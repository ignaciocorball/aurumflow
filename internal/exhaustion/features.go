package exhaustion

import (
	"math"
	"time"
)

type Trade struct {
	T          time.Time
	Price, Qty float64
	BuyerMaker bool
}

func (tr Trade) Notional() float64 { return tr.Price * tr.Qty }

type FlowWindow struct {
	Window           time.Duration `json:"window"`
	BuyVol           float64       `json:"aggressive_buy_volume"`
	SellVol          float64       `json:"aggressive_sell_volume"`
	BuyNotional      float64       `json:"aggressive_buy_notional"`
	SellNotional     float64       `json:"aggressive_sell_notional"`
	NetVol           float64       `json:"net_aggressive_volume"`
	NetNotional      float64       `json:"net_aggressive_notional"`
	TotalVol         float64       `json:"total_aggressive_volume"`
	TotalNotional    float64       `json:"total_aggressive_notional"`
	ImbalanceRatio   float64       `json:"flow_imbalance_ratio"`
	TradeCount       int           `json:"trade_count"`
	TradeVelocity    float64       `json:"trade_velocity"`
	NotionalVelocity float64       `json:"notional_velocity"`
	CVD              float64       `json:"cvd"`
	CVDDelta         float64       `json:"cvd_delta"`
	CVDSlope         float64       `json:"cvd_slope"`
	CVDAccel         float64       `json:"cvd_acceleration"`
}

type PriceWindow struct {
	Window       time.Duration `json:"window"`
	LogReturn    float64       `json:"log_return"`
	AbsReturn    float64       `json:"absolute_return"`
	Range        float64       `json:"range"`
	DispVsFlow   float64       `json:"displacement_vs_aggressive_flow"`
	DispVsLegacy float64       `json:"displacement_vs_legacy"`
}

type FeatureSet struct {
	At             time.Time              `json:"at"`
	Windows        map[string]FlowWindow  `json:"windows"`
	Prices         map[string]PriceWindow `json:"prices"`
	CVD            float64                `json:"cvd"`
	CVDPct         float64                `json:"cvd_percentile"`
	CVDZ           float64                `json:"cvd_z"`
	FlowMagNorm    float64                `json:"flow_magnitude_normalized"`
	PriceDispNorm  float64                `json:"price_displacement_normalized"`
	FlowEffNorm    float64                `json:"flow_efficiency_normalized"`
	ImpactFailure  float64                `json:"impact_failure"`
	RealizedVol    float64                `json:"realized_volatility"`
	ATR            float64                `json:"atr"`
	ATRPct         float64                `json:"atr_percentile"`
	IntensityPct   float64                `json:"trade_intensity_percentile"`
	UTCHour        int                    `json:"utc_hour"`
	Dow            int                    `json:"day_of_week"`
	Session        string                 `json:"session"`
	LegacyScore    int                    `json:"legacy_score"`
	LegacyDir      int                    `json:"legacy_direction"`
	DistInvalidATR float64                `json:"distance_to_invalidation_atr"`
	SweepState     string                 `json:"sweep_state,omitempty"`
	Equilibrium    string                 `json:"equilibrium_state,omitempty"`
}

func WindowKey(d time.Duration) string {
	return d.String()
}

func FlowFrom(trades []Trade, t0 time.Time, win time.Duration, cvdBefore float64) FlowWindow {
	start := t0.Add(-win)
	var w FlowWindow
	w.Window = win
	var firstCVD, lastCVD float64
	var have bool
	cvd := cvdBefore
	// rebuild cvd from trades before start
	for _, tr := range trades {
		if !tr.T.Before(start) {
			break
		}
		if tr.T.Before(t0) {
			if tr.BuyerMaker {
				cvd -= tr.Qty
			} else {
				cvd += tr.Qty
			}
		}
	}
	firstCVD = cvd
	mid := cvd
	midT := start.Add(win / 2)
	for _, tr := range trades {
		if tr.T.Before(start) || !tr.T.Before(t0) {
			continue
		}
		n := tr.Notional()
		w.TradeCount++
		w.TotalVol += tr.Qty
		w.TotalNotional += n
		if tr.BuyerMaker {
			w.SellVol += tr.Qty
			w.SellNotional += n
			cvd -= tr.Qty
		} else {
			w.BuyVol += tr.Qty
			w.BuyNotional += n
			cvd += tr.Qty
		}
		if !have && !tr.T.Before(midT) {
			mid = cvd
			have = true
		}
	}
	lastCVD = cvd
	w.NetVol = w.BuyVol - w.SellVol
	w.NetNotional = w.BuyNotional - w.SellNotional
	w.CVD = lastCVD
	w.CVDDelta = lastCVD - firstCVD
	sec := win.Seconds()
	if sec > 0 {
		w.CVDSlope = w.CVDDelta / sec
		w.TradeVelocity = float64(w.TradeCount) / sec
		w.NotionalVelocity = w.TotalNotional / sec
		w.CVDAccel = (lastCVD - mid - (mid - firstCVD)) / sec
	}
	den := w.TotalVol
	if den > 0 {
		w.ImbalanceRatio = w.NetVol / den
	}
	return w
}

func PriceFrom(trades []Trade, t0 time.Time, win time.Duration, legacyDir int, netFlow float64) PriceWindow {
	start := t0.Add(-win)
	var p PriceWindow
	p.Window = win
	var first, last, hi, lo float64
	n := 0
	for _, tr := range trades {
		if tr.T.Before(start) || !tr.T.Before(t0) {
			continue
		}
		if n == 0 {
			first, last, hi, lo = tr.Price, tr.Price, tr.Price, tr.Price
		}
		last = tr.Price
		if tr.Price > hi {
			hi = tr.Price
		}
		if tr.Price < lo {
			lo = tr.Price
		}
		n++
	}
	if n == 0 || first <= 0 {
		return p
	}
	p.LogReturn = math.Log(last / first)
	p.AbsReturn = math.Abs(p.LogReturn)
	p.Range = (hi - lo) / first
	flowSign := 1.0
	if netFlow < 0 {
		flowSign = -1
	} else if netFlow == 0 {
		flowSign = 0
	}
	p.DispVsFlow = p.LogReturn * flowSign
	p.DispVsLegacy = p.LogReturn * float64(legacyDir)
	return p
}

func PastOnly(trades []Trade, t0 time.Time) []Trade {
	out := trades[:0]
	for _, tr := range trades {
		if tr.T.Before(t0) {
			out = append(out, tr)
		}
	}
	return out
}

func RollingMedianMAD(xs []float64) (med, mad float64) {
	if len(xs) == 0 {
		return 0, 0
	}
	cp := append([]float64(nil), xs...)
	sortFloats(cp)
	med = quantile(cp, 0.5)
	dev := make([]float64, len(cp))
	for i, x := range cp {
		dev[i] = math.Abs(x - med)
	}
	sortFloats(dev)
	mad = quantile(dev, 0.5)
	return med, mad
}

func RobustZ(x, med, mad float64) float64 {
	if mad <= 1e-12 {
		return 0
	}
	return (x - med) / (1.4826 * mad)
}

func Percentile(xs []float64, v float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	n := 0
	for _, x := range xs {
		if x <= v {
			n++
		}
	}
	return 100 * float64(n) / float64(len(xs))
}

func ZScore(xs []float64, v float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	m := 0.0
	for _, x := range xs {
		m += x
	}
	m /= float64(len(xs))
	ss := 0.0
	for _, x := range xs {
		d := x - m
		ss += d * d
	}
	sd := math.Sqrt(ss / float64(len(xs)-1))
	if sd == 0 {
		return 0
	}
	return (v - m) / sd
}

func sortFloats(xs []float64) {
	for i := 1; i < len(xs); i++ {
		j := i
		for j > 0 && xs[j] < xs[j-1] {
			xs[j], xs[j-1] = xs[j-1], xs[j]
			j--
		}
	}
}

func quantile(sorted []float64, q float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if q <= 0 {
		return sorted[0]
	}
	if q >= 1 {
		return sorted[len(sorted)-1]
	}
	idx := q * float64(len(sorted)-1)
	i := int(idx)
	if i >= len(sorted)-1 {
		return sorted[len(sorted)-1]
	}
	f := idx - float64(i)
	return sorted[i]*(1-f) + sorted[i+1]*f
}
