package mktwarmup

import (
	"time"

	"aurumflow/internal/livesurface"
)

const (
	StatusReady       = "READY"
	StatusWarming     = "WARMING"
	StatusUnavailable = "HISTORY_UNAVAILABLE_WHILE_CLOSED"
	StatusStale       = "STALE"
	StatusUnknown     = "UNKNOWN"
)

type History struct {
	Market     string
	Status     string
	M5, M15, H1, H4 int
	Oldest     time.Time
	Latest     time.Time
	Gaps       int
	Stale      bool
	M5Bars     []livesurface.Candle
	M15Bars    []livesurface.Candle
	H1Bars     []livesurface.Candle
	H4Bars     []livesurface.Candle
}

func Assess(h History, now time.Time) History {
	if h.M5 == 0 && h.H1 == 0 && h.H4 == 0 {
		if h.Status == "" {
			h.Status = StatusUnknown
		}
		return h
	}
	if !h.Latest.IsZero() && now.Sub(h.Latest) > 36*time.Hour {
		h.Stale = true
		h.Status = StatusStale
		return h
	}
	if h.M5 >= 50 && h.H1 >= 20 && h.H4 >= 10 {
		h.Status = StatusReady
		return h
	}
	h.Status = StatusWarming
	return h
}

func FromBars(market string, m5, h1, h4 []livesurface.Candle, now time.Time) History {
	h := History{Market: market, M5Bars: m5, H1Bars: h1, H4Bars: h4, M5: len(m5), H1: len(h1), H4: len(h4)}
	var times []time.Time
	for _, xs := range [][]livesurface.Candle{m5, h1, h4} {
		for _, c := range xs {
			if !c.Time.IsZero() {
				times = append(times, c.Time)
			}
		}
	}
	for _, t := range times {
		if h.Oldest.IsZero() || t.Before(h.Oldest) {
			h.Oldest = t
		}
		if h.Latest.IsZero() || t.After(h.Latest) {
			h.Latest = t
		}
	}
	h.Gaps = countGaps(m5, 6*time.Minute)
	return Assess(h, now)
}

func countGaps(xs []livesurface.Candle, maxStep time.Duration) int {
	n := 0
	for i := 1; i < len(xs); i++ {
		if xs[i].Time.Sub(xs[i-1].Time) > maxStep {
			n++
		}
	}
	return n
}

func ClosedUnavailable(market string) History {
	return History{Market: market, Status: StatusUnavailable}
}
