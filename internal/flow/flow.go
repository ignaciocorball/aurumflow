package flow

import (
	"math"
	"time"
)

type Trade struct {
	Price, Qty float64
	BuyerMaker bool // true => taker sold (aggressive sell)
}

type Windows struct {
	Buy, Sell, Signed, Count, Notional float64
}

type Engine struct {
	cvd float64
}

func (e *Engine) OnTrade(tr Trade) (signed float64) {
	signed = tr.Qty
	if tr.BuyerMaker {
		signed = -tr.Qty
	}
	e.cvd += signed
	return signed
}

func (e *Engine) CVD() float64 { return e.cvd }

func ClassifyAggressor(buyerMaker bool) string {
	if buyerMaker {
		return "aggressive_sell"
	}
	return "aggressive_buy"
}

func Roll(trades []Trade) Windows {
	var w Windows
	for _, tr := range trades {
		notional := tr.Price * tr.Qty
		w.Count++
		w.Notional += notional
		if tr.BuyerMaker {
			w.Sell += tr.Qty
			w.Signed -= tr.Qty
		} else {
			w.Buy += tr.Qty
			w.Signed += tr.Qty
		}
	}
	return w
}

func ZScore(x, mean, std float64) float64 {
	if std <= 0 || math.IsNaN(std) {
		return 0
	}
	return (x - mean) / std
}

type TimedTrade struct {
	T time.Time
	Trade
}

func Window(trades []TimedTrade, now time.Time, d time.Duration) Windows {
	var subset []Trade
	for _, tr := range trades {
		if now.Sub(tr.T) <= d && !tr.T.After(now) {
			subset = append(subset, tr.Trade)
		}
	}
	return Roll(subset)
}

func Median(xs []float64) float64 {
	if len(xs) == 0 {
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
	if len(cp)%2 == 1 {
		return cp[len(cp)/2]
	}
	return (cp[len(cp)/2-1] + cp[len(cp)/2]) / 2
}
