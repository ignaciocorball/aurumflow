package finra

import (
	"math"
	"sort"
)

type Week struct {
	Symbol     string
	WeekEnding timeWeek
	ATS        float64
	NonATS     float64
	Total      float64
	Present    bool
}

type timeWeek = struct {
	Year  int
	Month int
	Day   int
	Raw   string
}

func ParseWeek(symbol, ending string, ats, nonATS float64, present bool) Week {
	return Week{Symbol: symbol, WeekEnding: timeWeek{Raw: ending}, ATS: ats, NonATS: nonATS, Total: ats + nonATS, Present: present}
}

func Feature(weeks []Week, symbol, ending string) (ats, nonATS, total, change, pct, z float64, missing bool) {
	var hist []float64
	var cur *Week
	for i := range weeks {
		w := weeks[i]
		if w.Symbol != symbol {
			continue
		}
		if !w.Present {
			continue
		}
		hist = append(hist, w.Total)
		if w.WeekEnding.Raw == ending {
			cur = &weeks[i]
		}
	}
	if cur == nil {
		return 0, 0, 0, 0, 0, 0, true
	}
	ats, nonATS, total = cur.ATS, cur.NonATS, cur.Total
	if len(hist) >= 2 {
		change = total - hist[len(hist)-2]
	}
	pct = percentile(hist, total)
	z = zscore(hist, total)
	return ats, nonATS, total, change, pct, z, false
}

func percentile(xs []float64, v float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	cp := append([]float64(nil), xs...)
	sort.Float64s(cp)
	n := 0
	for _, x := range cp {
		if x <= v {
			n++
		}
	}
	return 100 * float64(n) / float64(len(cp))
}

func zscore(xs []float64, v float64) float64 {
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
