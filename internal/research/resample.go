package research

import (
	"time"

	"aurumflow/pkg/models"
)

func Resample(in []models.Candle, d time.Duration) []models.Candle {
	if d <= 0 || len(in) == 0 {
		return nil
	}
	var out []models.Candle
	var cur models.Candle
	var bucket time.Time
	for _, c := range in {
		b := c.Time.UTC().Truncate(d)
		if bucket.IsZero() {
			bucket = b
			cur = models.Candle{Time: b, Open: c.Open, High: c.High, Low: c.Low, Close: c.Close, Volume: c.Volume}
			continue
		}
		if !b.Equal(bucket) {
			out = append(out, cur)
			bucket = b
			cur = models.Candle{Time: b, Open: c.Open, High: c.High, Low: c.Low, Close: c.Close, Volume: c.Volume}
			continue
		}
		if c.High > cur.High {
			cur.High = c.High
		}
		if c.Low < cur.Low {
			cur.Low = c.Low
		}
		cur.Close = c.Close
		cur.Volume += c.Volume
	}
	if !bucket.IsZero() {
		out = append(out, cur)
	}
	return out
}

func CandlePrices(cs []models.Candle) []PricePoint {
	out := make([]PricePoint, 0, len(cs))
	for _, c := range cs {
		out = append(out, PricePoint{T: c.Time, P: c.Close})
	}
	return out
}
