package strathist

import (
	"time"

	"aurumflow/internal/capitalhist"
	"aurumflow/pkg/models"
)

func ClosedOnly(cs []models.Candle, asOf time.Time, period time.Duration) []models.Candle {
	out := make([]models.Candle, 0, len(cs))
	for _, c := range cs {
		if RejectLookahead(c, asOf, period) {
			continue
		}
		if Incomplete(c, asOf, period) {
			continue
		}
		if !ValidOHLC(c) {
			continue
		}
		out = append(out, c)
	}
	return capitalhist.DedupSort(out)
}

func RejectLookahead(c models.Candle, asOf time.Time, period time.Duration) bool {
	end := c.Time
	if period > 0 {
		end = c.Time.Add(period)
	}
	return !end.IsZero() && end.After(asOf)
}

func Incomplete(c models.Candle, asOf time.Time, period time.Duration) bool {
	if period <= 0 {
		return c.Time.After(asOf)
	}
	return c.Time.Add(period).After(asOf)
}

func ValidOHLC(c models.Candle) bool {
	if c.Time.IsZero() || c.Open <= 0 || c.High <= 0 || c.Low <= 0 || c.Close <= 0 {
		return false
	}
	if c.High < c.Low {
		return false
	}
	if c.High < c.Open || c.High < c.Close || c.Low > c.Open || c.Low > c.Close {
		return false
	}
	return true
}

func Merge(seed, live []models.Candle) []models.Candle {
	byTime := make(map[int64]models.Candle, len(seed)+len(live))
	for _, c := range seed {
		byTime[c.Time.UnixNano()] = c
	}
	for _, c := range live {
		byTime[c.Time.UnixNano()] = c
	}
	out := make([]models.Candle, 0, len(byTime))
	for _, c := range byTime {
		out = append(out, c)
	}
	return capitalhist.DedupSort(out)
}

func ExpectedWeekendGap(prev, next time.Time) bool {
	if next.Before(prev) || next.Equal(prev) {
		return false
	}
	p, n := prev.UTC(), next.UTC()
	if n.Sub(p) < 12*time.Hour {
		return false
	}
	if p.Weekday() == time.Friday && (n.Weekday() == time.Sunday || n.Weekday() == time.Monday) {
		return true
	}
	if p.Weekday() == time.Saturday && (n.Weekday() == time.Sunday || n.Weekday() == time.Monday) {
		return true
	}
	return false
}

func ClassifyGaps(cs []models.Candle, step time.Duration) (expected, unexpected int) {
	if step <= 0 || len(cs) < 2 {
		return 0, 0
	}
	for i := 1; i < len(cs); i++ {
		delta := cs[i].Time.Sub(cs[i-1].Time)
		if delta <= step+time.Minute {
			continue
		}
		if ExpectedWeekendGap(cs[i-1].Time, cs[i].Time) {
			expected++
			continue
		}
		unexpected++
	}
	return expected, unexpected
}

func OldestLatest(cs []models.Candle) (time.Time, time.Time) {
	if len(cs) == 0 {
		return time.Time{}, time.Time{}
	}
	return cs[0].Time, cs[len(cs)-1].Time
}
