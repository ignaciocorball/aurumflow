package legacynorm

import (
	"math"
	"sort"
	"time"

	"aurumflow/internal/backtest"
	"aurumflow/internal/indicators"
	"aurumflow/internal/research"
	"aurumflow/pkg/models"
)

const Spec = "LEGACY_NORMALIZED_V0"
const SpecHash = "sha256:f4c22591fbb0086c3a0cde0d49f1f54f5b5d97b30bfc1c811c563bd59d78387f"

func CurrentLegacyConfig() backtest.Config {
	cfg := research.CryptoResearchConfig()
	cfg.Crypto = false
	cfg.MinATR = 8
	cfg.MaxATR = 18
	cfg.TradingSessions = []string{"LONDON", "NY"}
	cfg.BlockH1Range = true
	return cfg
}

func NormalizedConfig() backtest.Config {
	cfg := research.CryptoResearchConfig()
	cfg.Crypto = false
	cfg.MinATR = 0
	cfg.MaxATR = 1e9
	cfg.TradingSessions = []string{"ALL"}
	cfg.BlockH1Range = true
	return cfg
}

func SessionsOf(market string) []string {
	switch market {
	case "US100", "US500", "US30":
		return []string{"NY"}
	case "DE40", "UK100":
		return []string{"LONDON"}
	case "J225", "CN50":
		return []string{"ASIA"}
	default:
		return []string{"ALL"}
	}
}

func Compat(market string, medianATR float64) string {
	if market == "BTC" {
		return "UNSUPPORTED"
	}
	if medianATR <= 0 {
		return "INSUFFICIENT_DATA"
	}
	if market == "GOLD" {
		return "MECHANICALLY_COMPATIBLE"
	}
	if medianATR < 8 || medianATR > 18 {
		if market == "J225" || market == "CN50" {
			return "SESSION_INCOMPATIBLE"
		}
		return "SCALE_INCOMPATIBLE"
	}
	if market == "J225" || market == "CN50" {
		return "SESSION_INCOMPATIBLE"
	}
	return "MECHANICALLY_COMPATIBLE"
}

func ScanNormalized(m15, h1, h4 []models.Candle) []research.SignalRow {
	raw := research.ScanLegacy(m15, h1, h4, NormalizedConfig())
	var out []research.SignalRow
	for _, s := range raw {
		idx := indexAt(m15, s.Time)
		if idx < 54 {
			continue
		}
		if !VolSuitable(m15[:idx+1]) {
			continue
		}
		out = append(out, s)
	}
	return out
}

func indexAt(cs []models.Candle, t time.Time) int {
	for i := len(cs) - 1; i >= 0; i-- {
		if !cs[i].Time.After(t) {
			return i
		}
	}
	return -1
}

func VolSuitable(past []models.Candle) bool {
	rank, ok := ATRPctRank(past)
	if !ok {
		return false
	}
	return rank >= 0.20 && rank <= 0.80
}

func ATRPctRank(past []models.Candle) (float64, bool) {
	if len(past) < 55 {
		return 0, false
	}
	cur := atrPct(past)
	if cur <= 0 {
		return 0, false
	}
	look := 200
	if len(past)-15 < look {
		look = len(past) - 15
	}
	if look < 40 {
		return 0, false
	}
	var hist []float64
	for i := 15; i < look+15 && i < len(past)-1; i++ {
		v := atrPct(past[: i+1])
		if v > 0 {
			hist = append(hist, v)
		}
	}
	if len(hist) < 40 {
		return 0, false
	}
	sort.Float64s(hist)
	below := 0
	for _, v := range hist {
		if v <= cur {
			below++
		}
	}
	return float64(below) / float64(len(hist)), true
}

func atrPct(cs []models.Candle) float64 {
	atr := indicators.ATR(cs, 14)
	px := cs[len(cs)-1].Close
	if px <= 0 || math.IsNaN(atr) || atr <= 0 {
		return 0
	}
	return atr / px
}

func MedianATR(cs []models.Candle) float64 {
	var xs []float64
	for i := 20; i < len(cs); i++ {
		v := indicators.ATR(cs[:i+1], 14)
		if !math.IsNaN(v) && v > 0 {
			xs = append(xs, v)
		}
	}
	if len(xs) == 0 {
		return 0
	}
	sort.Float64s(xs)
	return xs[len(xs)/2]
}
