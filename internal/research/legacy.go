package research

import (
	"time"

	"aurumflow/internal/backtest"
	"aurumflow/internal/core"
	"aurumflow/internal/marketstate"
	"aurumflow/internal/strategy"
	"aurumflow/pkg/models"
)

type SignalRow struct {
	Time              time.Time
	Direction         int
	Score             int
	State             string
	Source            string
	Pressure          float64
	OriginalPressure  float64
	DirPressure       float64
	FlowInterp        string
	Confidence        float64
	RadarState        string
	Class             string
	CVD               float64
	AggBuy            float64
	AggSell           float64
	PreReturn         float64
	TradeVel          float64
}

func CryptoResearchConfig() backtest.Config {
	return backtest.Config{
		RSIPeriod: 14, ATRPeriod: 14, MinATR: 0, MaxATR: 1e9,
		SwingLookback: 5, ScoreThreshold: 5, RSIRequired: false, RSIMode: "BOOST",
		RSIBuyLow: 38, RSIBuyHigh: 42, RSISellLow: 58, RSISellHigh: 62,
		SweepCloseTolerance: 0.15, SweepMinPenetration: 0.05, SweepMinWickRatio: 0.4,
		UseH1Filter: true, BlockH1Range: true, Crypto: true, H1RangeExtraScore: 1,
		H1SwingLookback: 5, UseH4Filter: true, BlockH4Range: false, H4SwingLookback: 5,
		StateTTLCandles: 12, BlockLondonNYOverlap: false, TradingSessions: []string{"ALL"},
		SpreadPoints: 0, SlippagePoints: 0,
	}
}

func ScanLegacy(m15, h1, h4 []models.Candle, cfg backtest.Config) []SignalRow {
	if cfg.ScoreThreshold == 0 {
		cfg = CryptoResearchConfig()
	}
	sm := core.NewStateMachine()
	lookback := cfg.SwingLookback
	if lookback <= 0 {
		lookback = 5
	}
	sweepOpts := &strategy.SweepOptions{
		CloseTolerance: cfg.SweepCloseTolerance, MinPenetration: cfg.SweepMinPenetration, MinWickRatio: cfg.SweepMinWickRatio,
	}
	var out []SignalRow
	h1Idx, h4Idx := 0, 0
	for i := 20; i < len(m15); i++ {
		slice := m15[: i+1]
		in := strategy.BuildComposerInput(slice, lookback, cfg.EntryDelayCandles, cfg.RSIPeriod, cfg.ATRPeriod, sweepOpts, cfg.UseRSIWilder)
		setupDirection := ""
		if in.LiquidityEvent != nil {
			if in.LiquidityEvent.Type == "BUY_SIDE" {
				setupDirection = "BUY"
			} else {
				setupDirection = "SELL"
			}
		} else if in.SweepBuy {
			setupDirection = "BUY"
		} else if in.SweepSell {
			setupDirection = "SELL"
		}
		sm.Transition(len(in.LiquidityZones) > 0, strategy.HasSweep(in), in.EquilibriumTouch, false, core.TransitionParams{
			LastCandleTime: m15[i].Time, StructureTrend: in.Structure.Trend, SetupDirection: setupDirection,
			StateTTLCandles: cfg.StateTTLCandles, InvalidateOnOppositeStructure: cfg.InvalidateOnOppositeStructure,
		})
		for h1Idx < len(h1) && !h1[h1Idx].Time.After(m15[i].Time) {
			h1Idx++
		}
		for h4Idx < len(h4) && !h4[h4Idx].Time.After(m15[i].Time) {
			h4Idx++
		}
		if sm.State() != core.StateReady {
			continue
		}
		sig, ok := strategy.SignalComposerWithZones(in, cfg.MinATR, cfg.MaxATR, cfg.ScoreThreshold,
			cfg.RSIBuyLow, cfg.RSIBuyHigh, cfg.RSISellLow, cfg.RSISellHigh, cfg.RSIRequired, cfg.RSIMode)
		if !ok || sig == nil {
			continue
		}
		trendH1 := ""
		if h1Idx >= 10 {
			trendH1 = marketstate.TrendFromCandles(h1[:h1Idx], 5)
		}
		if cfg.UseH1Filter && trendH1 == "RANGE" && cfg.BlockH1Range && !cfg.Crypto {
			continue
		}
		dir := 1
		if sig.Direction == "SELL" {
			dir = -1
		}
		out = append(out, SignalRow{Time: m15[i].Time, Direction: dir, Score: sig.Score, State: string(sm.State()), Source: "legacy"})
	}
	return out
}
