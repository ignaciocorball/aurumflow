package market

import "fmt"

// InstrumentSpec is normalized dealing-rule data. Missing critical fields stay zero;
// callers must not invent values.
type InstrumentSpec struct {
	Epic           string
	InstrumentType string
	Name           string
	MinDealSize    float64
	MaxDealSize    float64
	SizeStep       float64
	LotSize        float64
	ValuePerPoint  float64 // from config only; 0 = unknown
}

// SizingComplete is true when broker min/max/step and value-per-point are all usable.
func (s InstrumentSpec) SizingComplete() bool {
	return s.Epic != "" && s.MinDealSize > 0 && s.MaxDealSize > 0 && s.SizeStep > 0 && s.ValuePerPoint > 0 && s.MaxDealSize >= s.MinDealSize
}

// SpecFromDetails maps Capital.com market details. valuePerPoint must come from config (not invented).
func SpecFromDetails(details *MarketDetailsResponse, valuePerPoint float64) (InstrumentSpec, error) {
	if details == nil {
		return InstrumentSpec{}, fmt.Errorf("INSTRUMENT_SPEC_INCOMPLETE: market details missing")
	}
	spec := InstrumentSpec{
		Epic:           details.Instrument.Epic,
		InstrumentType: details.Instrument.Type,
		Name:           details.Instrument.Name,
		MinDealSize:    details.DealingRules.MinDealSize.Value,
		MaxDealSize:    details.DealingRules.MaxDealSize.Value,
		SizeStep:       details.DealingRules.MinSizeIncrement.Value,
		LotSize:        details.Instrument.LotSize,
		ValuePerPoint:  valuePerPoint,
	}
	if spec.Epic == "" {
		return spec, fmt.Errorf("INSTRUMENT_SPEC_INCOMPLETE: epic missing")
	}
	return spec, nil
}
