package pilotrisk

import (
	"math"
	"strings"

	"aurumflow/internal/market"
	"aurumflow/internal/portfoliorisk"
	"aurumflow/internal/risk"
)

const (
	Policy           = "PILOT_DEMO_RISK_V1"
	PerPositionPct   = 0.50
	AggregatePct     = 1.00
	MaxTradesMarket  = 1
	MaxStrategyOpen  = 2
	HardAggregateUSD = 300.0
	HardClassUSD     = 150.0

	ReasonMinSize     = "MIN_SIZE_EXCEEDS_PILOT_RISK"
	ReasonPerPosition = "PILOT_PER_POSITION_CAP"
	ReasonAggregate   = "PILOT_AGGREGATE_CAP"
	ReasonTradeLimit  = "PILOT_TRADE_LIMIT"
	ReasonMaxOpen     = "PILOT_MAX_POSITIONS"
	ReasonGroup       = "PILOT_CORRELATION_GROUP"
	ReasonEquity      = "PILOT_EQUITY_UNKNOWN"
	ReasonStop        = "PILOT_STOP_UNKNOWN"
)

// Policy V1 is an operations overlay. It does not change entry, SL, TP, or sessions.

type Input struct {
	Market           string
	Equity           float64
	StopDistance     float64
	MPU              float64
	MinDealSize      float64
	SizeStep         float64
	MaxDealSize      float64
	OpenRisk         float64
	TradesThisMarket int
	OpenStrategy     int
	GroupOpen        map[string]int
	HardAggregate    float64
	HardClass        float64
}

type Decision struct {
	Pass          bool
	Block         string
	Size          float64
	PlannedRisk   float64
	PerPosCap     float64
	AggregateCap  float64
	RemainingAgg  float64
	EffectivePct  float64
}

func PerPositionCap(equity float64) float64 {
	if equity <= 0 {
		return 0
	}
	return equity * PerPositionPct / 100
}

func AggregateCap(equity float64) float64 {
	if equity <= 0 {
		return 0
	}
	return equity * AggregatePct / 100
}

func EffectivePerPosition(equity, hardClass float64) float64 {
	if hardClass <= 0 {
		hardClass = HardClassUSD
	}
	return math.Min(PerPositionCap(equity), hardClass)
}

func EffectiveAggregate(equity, hardAgg float64) float64 {
	if hardAgg <= 0 {
		hardAgg = HardAggregateUSD
	}
	return math.Min(AggregateCap(equity), hardAgg)
}

func Evaluate(in Input) Decision {
	d := Decision{}
	if in.HardClass <= 0 {
		in.HardClass = HardClassUSD
	}
	if in.HardAggregate <= 0 {
		in.HardAggregate = HardAggregateUSD
	}
	d.PerPosCap = EffectivePerPosition(in.Equity, in.HardClass)
	d.AggregateCap = EffectiveAggregate(in.Equity, in.HardAggregate)
	d.RemainingAgg = d.AggregateCap - in.OpenRisk
	if in.Equity <= 0 {
		d.Block = ReasonEquity
		return d
	}
	if in.StopDistance <= 0 || in.MPU <= 0 {
		d.Block = ReasonStop
		return d
	}
	if in.TradesThisMarket >= MaxTradesMarket {
		d.Block = ReasonTradeLimit
		return d
	}
	if in.OpenStrategy >= MaxStrategyOpen {
		d.Block = ReasonMaxOpen
		return d
	}
	g := portfoliorisk.GroupOf(in.Market)
	if g != "" && in.GroupOpen[g] >= 1 {
		d.Block = ReasonGroup + " " + g
		return d
	}
	if d.RemainingAgg <= 0 {
		d.Block = ReasonAggregate
		return d
	}
	budget := math.Min(d.PerPosCap, d.RemainingAgg)
	minRisk := in.MinDealSize * in.StopDistance * in.MPU
	if in.MinDealSize > 0 && minRisk > budget+1e-9 {
		d.Block = ReasonMinSize
		d.PlannedRisk = minRisk
		return d
	}
	spec := market.InstrumentSpec{
		Epic: strings.ToUpper(in.Market), MinDealSize: in.MinDealSize, MaxDealSize: in.MaxDealSize,
		SizeStep: in.SizeStep, ValuePerPoint: in.MPU,
	}
	if spec.MaxDealSize <= 0 {
		spec.MaxDealSize = 1e9
	}
	if spec.SizeStep <= 0 {
		spec.SizeStep = in.MinDealSize
	}
	d.EffectivePct = budget / in.Equity * 100
	size, err := risk.ComputeSize(in.Equity, d.EffectivePct, in.StopDistance, spec)
	if err != nil {
		if strings.Contains(err.Error(), risk.ReasonMinSizeExceedsRisk) {
			d.Block = ReasonMinSize
			d.PlannedRisk = minRisk
			return d
		}
		d.Block = err.Error()
		return d
	}
	d.Size = size
	d.PlannedRisk = size * in.StopDistance * in.MPU
	if d.PlannedRisk > d.PerPosCap+1e-6 {
		d.Block = ReasonPerPosition
		return d
	}
	if in.OpenRisk+d.PlannedRisk > d.AggregateCap+1e-6 {
		d.Block = ReasonAggregate
		return d
	}
	d.Pass = true
	return d
}

func ModeledRisk(size, stopDistance, mpu float64) float64 {
	if size <= 0 || stopDistance <= 0 || mpu <= 0 {
		return 0
	}
	return size * stopDistance * mpu
}
