package research

import (
	"fmt"
	"sort"
	"time"

	"aurumflow/internal/radar"
)

type RadarPoint struct {
	Time       time.Time
	Pressure   float64
	Confidence float64
	Direction  int
	State      string
	CVD        float64
	AggBuy     float64
	AggSell    float64
	TradeVel   float64
}

type Study struct {
	Name   string
	Stats  map[string]BucketStats // horizon -> stats
	N      int
}

func AlignFusion(legacy []SignalRow, radar []RadarPoint, alignAbs float64) []SignalRow {
	out := make([]SignalRow, 0, len(legacy))
	for _, s := range legacy {
		rp := latestRadar(radar, s.Time)
		class := Classify(s.Direction, rp.Direction, rp.Pressure, alignAbs)
		s.Pressure = rp.Pressure
		s.OriginalPressure = rp.Pressure
		s.DirPressure = DirectionalPressure(s.Direction, rp.Pressure)
		s.FlowInterp = ClassifyFlow(s.Direction, rp.Pressure, alignAbs)
		s.Confidence = rp.Confidence
		s.RadarState = rp.State
		s.Class = class
		s.CVD = rp.CVD
		s.AggBuy = rp.AggBuy
		s.AggSell = rp.AggSell
		s.TradeVel = rp.TradeVel
		out = append(out, s)
	}
	return out
}

func latestRadar(radar []RadarPoint, at time.Time) RadarPoint {
	// last point with Time <= at
	i := sort.Search(len(radar), func(i int) bool { return radar[i].Time.After(at) })
	if i == 0 {
		return RadarPoint{}
	}
	return radar[i-1]
}

func StudySignals(rows []SignalRow, prices []PricePoint, horizon time.Duration, keep func(SignalRow) bool) BucketStats {
	var rets, mfes, maes []float64
	for _, r := range rows {
		if keep != nil && !keep(r) {
			continue
		}
		ls := Labels(prices, r.Time, priceAt(prices, r.Time), r.Direction, []time.Duration{horizon})
		if len(ls) == 0 || !ls[0].Complete && ls[0].ForwardRet == 0 {
			continue
		}
		rets = append(rets, ls[0].DirRet)
		mfes = append(mfes, ls[0].MFE)
		maes = append(maes, ls[0].MAE)
	}
	s := Summarize(rets, mfes, maes)
	s.Name = horizon.String()
	return s
}

func priceAt(prices []PricePoint, at time.Time) float64 {
	i := sort.Search(len(prices), func(i int) bool { return prices[i].T.After(at) })
	if i == 0 {
		if len(prices) == 0 {
			return 0
		}
		return prices[0].P
	}
	return prices[i-1].P
}

func PressureBuckets(rows []SignalRow, prices []PricePoint, horizon time.Duration) []BucketStats {
	bounds := [][2]float64{{0, 20}, {20, 40}, {40, 60}, {60, 80}, {80, 101}}
	var out []BucketStats
	for _, b := range bounds {
		lo, hi := b[0], b[1]
		st := StudySignals(rows, prices, horizon, func(r SignalRow) bool {
			a := abs(r.Pressure)
			return a >= lo && a < hi
		})
		st.Name = formatBucket(lo, hi)
		out = append(out, st)
	}
	return out
}

func ConfidenceBuckets(rows []SignalRow, prices []PricePoint, horizon time.Duration) []BucketStats {
	bounds := [][2]float64{{0, 40}, {40, 60}, {60, 80}, {80, 101}}
	var out []BucketStats
	for _, b := range bounds {
		st := StudySignals(rows, prices, horizon, func(r SignalRow) bool {
			return r.Confidence >= b[0] && r.Confidence < b[1]
		})
		st.Name = formatBucket(b[0], b[1])
		out = append(out, st)
	}
	return out
}

func StateBuckets(rows []SignalRow, prices []PricePoint, horizon time.Duration) []BucketStats {
	var out []BucketStats
	for _, name := range []string{radar.StateNoTrade, radar.StateAbsorption, radar.StateExpansion} {
		st := StudySignals(rows, prices, horizon, func(r SignalRow) bool { return r.RadarState == name })
		st.Name = name
		out = append(out, st)
	}
	return out
}

func FilterValue(legacy []SignalRow, prices []PricePoint, horizon time.Duration) (all, filtered BucketStats) {
	all = StudySignals(legacy, prices, horizon, nil)
	filtered = StudySignals(legacy, prices, horizon, func(r SignalRow) bool {
		return r.Class != ContradictedLong && r.Class != ContradictedShort
	})
	return all, filtered
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func formatBucket(lo, hi float64) string {
	return fmt.Sprintf("%.0f-%.0f", lo, hi)
}

func ToRadarPoints(snaps []radar.PressureSnapshot) []RadarPoint {
	out := make([]RadarPoint, 0, len(snaps))
	for _, s := range snaps {
		out = append(out, RadarPoint{Time: s.Timestamp, Pressure: s.PressureScore, Confidence: s.Confidence, Direction: s.Direction, State: s.State, CVD: s.CVD, AggBuy: s.AggBuy, AggSell: s.AggSell, TradeVel: s.TradeVel})
	}
	return out
}
