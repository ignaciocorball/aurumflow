package candles

import (
	"time"

	"aurumflow/pkg/models"
)

type Trade struct {
	T     time.Time
	Price float64
	Qty   float64
}

func Synthesize(trades []Trade, d time.Duration) []models.Candle {
	if d <= 0 || len(trades) == 0 {
		return nil
	}
	var out []models.Candle
	var cur models.Candle
	var bucket time.Time
	for _, tr := range trades {
		if tr.Price <= 0 {
			continue
		}
		b := tr.T.UTC().Truncate(d)
		if bucket.IsZero() {
			bucket = b
			cur = models.Candle{Time: b, Open: tr.Price, High: tr.Price, Low: tr.Price, Close: tr.Price, Volume: tr.Qty}
			continue
		}
		if !b.Equal(bucket) {
			out = append(out, cur)
			bucket = b
			cur = models.Candle{Time: b, Open: tr.Price, High: tr.Price, Low: tr.Price, Close: tr.Price, Volume: tr.Qty}
			continue
		}
		if tr.Price > cur.High {
			cur.High = tr.Price
		}
		if tr.Price < cur.Low {
			cur.Low = tr.Price
		}
		cur.Close = tr.Price
		cur.Volume += tr.Qty
	}
	if !bucket.IsZero() {
		out = append(out, cur)
	}
	return out
}
