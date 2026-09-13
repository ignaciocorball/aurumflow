package radar

import "time"

const (
	StateNoTrade    = "NO_TRADE"
	StateAbsorption = "ABSORPTION"
	StateExpansion  = "EXPANSION"
	ModeOff         = "OFF"
	ModeShadow      = "SHADOW"

	CapTradeFlow  uint32 = 1 << iota
	CapBook
	CapCrossAsset
	CapContext
)

type Evidence struct {
	Name   string
	Points float64
	Note   string
}

type PressureSnapshot struct {
	Instrument string
	Timestamp  time.Time
	Direction  int
	PressureScore float64
	AggressiveFlowScore float64
	AbsorptionScore     float64
	BookImbalanceScore  float64
	LiquidityScore      float64
	PersistenceScore    float64
	VolatilityContext   float64
	State      string
	Confidence     float64
	Evidence       []Evidence
	BookSynced     bool
	TradeFlowOnly  bool
	Caps           uint32
	CVD            float64
	AggBuy         float64
	AggSell        float64
	TradeVel       float64
}

func Compose(in PressureSnapshot) PressureSnapshot {
	in.Evidence = []Evidence{
		{"AggressiveFlow", in.AggressiveFlowScore, ""},
		{"Absorption", in.AbsorptionScore, "POTENTIAL_ABSORPTION only"},
		{"BookImbalance", in.BookImbalanceScore, ""},
		{"Liquidity", in.LiquidityScore, ""},
		{"Persistence", in.PersistenceScore, ""},
		{"Volatility", in.VolatilityContext, ""},
	}
	in.PressureScore = in.AggressiveFlowScore + in.AbsorptionScore + in.BookImbalanceScore + in.LiquidityScore + in.PersistenceScore + in.VolatilityContext
	if in.PressureScore > 100 {
		in.PressureScore = 100
	}
	if in.PressureScore < -100 {
		in.PressureScore = -100
	}
	switch {
	case in.PressureScore > 15:
		in.Direction = 1
	case in.PressureScore < -15:
		in.Direction = -1
	default:
		in.Direction = 0
	}
	if !in.BookSynced && !in.TradeFlowOnly {
		in.Confidence *= 0.25
		in.State = StateNoTrade
	}
	if in.State == "" {
		if in.AbsorptionScore > 20 {
			in.State = StateAbsorption
		} else if in.PressureScore > 25 || in.PressureScore < -25 {
			in.State = StateExpansion
		} else {
			in.State = StateNoTrade
		}
	}
	return in
}

type Hysteresis struct {
	Last      string
	ChangedAt time.Time
	MinHold   time.Duration
}

func (h *Hysteresis) Allow(next string, now time.Time) bool {
	if h.Last == "" || h.Last == next {
		h.Last = next
		if h.ChangedAt.IsZero() {
			h.ChangedAt = now
		}
		return true
	}
	if now.Sub(h.ChangedAt) < h.MinHold {
		return false
	}
	h.Last = next
	h.ChangedAt = now
	return true
}
