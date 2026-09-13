package research

import (
	"math"
	"time"
)

type Horizon time.Duration

var DefaultHorizons = []time.Duration{
	10 * time.Second, 30 * time.Second,
	time.Minute, 3 * time.Minute, 5 * time.Minute,
	15 * time.Minute, 30 * time.Minute, time.Hour,
}

type PricePoint struct {
	T time.Time
	P float64
}

type Label struct {
	Horizon      time.Duration
	ForwardRet   float64
	DirRet       float64
	MFE          float64
	MAE          float64
	TimeToMFE    time.Duration
	TimeToMAE    time.Duration
	Complete     bool
}

func Labels(prices []PricePoint, at time.Time, entry float64, direction int, horizons []time.Duration) []Label {
	if direction == 0 {
		direction = 1
	}
	out := make([]Label, 0, len(horizons))
	for _, h := range horizons {
		out = append(out, labelOne(prices, at, entry, direction, h))
	}
	return out
}

func labelOne(prices []PricePoint, at time.Time, entry float64, direction int, h time.Duration) Label {
	end := at.Add(h)
	var mfe, mae float64
	var tMFE, tMAE time.Duration
	var last float64
	complete := false
	for _, px := range prices {
		if px.T.Before(at) || px.T.Equal(at) {
			continue
		}
		if px.T.After(end) {
			complete = true
			break
		}
		r := (px.P - entry) / entry
		if direction < 0 {
			r = -r
		}
		last = r
		if r > mfe {
			mfe = r
			tMFE = px.T.Sub(at)
		}
		if r < mae {
			mae = r
			tMAE = px.T.Sub(at)
		}
		complete = !px.T.Before(end)
	}
	return Label{Horizon: h, ForwardRet: last, DirRet: last, MFE: mfe, MAE: mae, TimeToMFE: tMFE, TimeToMAE: tMAE, Complete: complete || last != 0}
}

func MFEMAERatio(mfe, mae float64) float64 {
	if mae == 0 {
		if mfe == 0 {
			return 0
		}
		return 99
	}
	return math.Abs(mfe) / math.Abs(mae)
}
