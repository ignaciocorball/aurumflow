package backtest

import (
	"testing"
	"time"

	"aurumflow/pkg/models"
)

func TestRun(t *testing.T) {
	candles := make([]models.Candle, 200)
	base := 2000.0
	for i := range candles {
		candles[i] = models.Candle{
			Time:   time.Unix(int64(i)*3600, 0),
			Open:   base,
			High:   base + 5,
			Low:    base - 5,
			Close:  base,
			Volume: 100,
		}
		if i%20 < 10 {
			base += 2
		} else {
			base -= 1
		}
		candles[i].Close = base
	}
	cfg := Config{
		RSIPeriod:          9,
		ATRPeriod:          14,
		MinATR:             80,
		MaxATR:             350,
		SwingLookback:      5,
		EntryDelayCandles:  3,
		ScoreThreshold:     6,
		RiskPerTrade:       0.5,
		RSIRequired:        false,
		RSIBuyLow:          38,
		RSIBuyHigh:         42,
		RSISellLow:         58,
		RSISellHigh:        62,
		SweepCloseTolerance: 0.2,
		SweepMinPenetration: 0.12,
		SweepMinWickRatio:   1.4,
		UseH1Filter:         false,
		H1SwingLookback:     5,
	}
	signals, res := Run(candles, nil, nil, cfg, 10000, nil, "")
	_ = signals
	if res.TotalTrades < 0 {
		t.Error("total trades should be >= 0")
	}
	if res.MaxDrawdownPct < 0 {
		t.Error("max drawdown pct should be >= 0")
	}
}
