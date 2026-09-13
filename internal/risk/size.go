package risk

import (
	"fmt"
	"math"

	"aurumflow/internal/market"
)

const (
	ReasonMinSizeExceedsRisk = "MIN_SIZE_EXCEEDS_RISK_BUDGET"
	ReasonInstrumentSpec     = "INSTRUMENT_SPEC_INCOMPLETE"
	ReasonMaxSizeExceeded    = "MAX_SIZE_EXCEEDED"
)

// ComputeSize sizes a trade from risk budget and broker dealing rules.
// It never raises size to min (that would increase risk). It never rounds above max.
func ComputeSize(balance, riskPercent, stopDistance float64, spec market.InstrumentSpec) (float64, error) {
	if !spec.SizingComplete() {
		return 0, fmt.Errorf(ReasonInstrumentSpec)
	}
	if math.IsNaN(balance) || math.IsInf(balance, 0) || balance <= 0 {
		return 0, fmt.Errorf("invalid balance")
	}
	if math.IsNaN(riskPercent) || math.IsInf(riskPercent, 0) || riskPercent <= 0 {
		return 0, fmt.Errorf("invalid risk percent")
	}
	if math.IsNaN(stopDistance) || math.IsInf(stopDistance, 0) || stopDistance <= 0 {
		return 0, fmt.Errorf("invalid stop distance")
	}
	riskAmount := balance * (riskPercent / 100)
	raw := riskAmount / (math.Abs(stopDistance) * spec.ValuePerPoint)
	size := math.Floor(raw/spec.SizeStep) * spec.SizeStep
	if size+1e-12 < spec.MinDealSize {
		return 0, fmt.Errorf("%s: calculated=%.6f min=%.6f", ReasonMinSizeExceedsRisk, size, spec.MinDealSize)
	}
	if size > spec.MaxDealSize+1e-12 {
		return 0, fmt.Errorf("%s: calculated=%.6f max=%.6f", ReasonMaxSizeExceeded, size, spec.MaxDealSize)
	}
	return size, nil
}
