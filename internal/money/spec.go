package money

import (
	"fmt"
	"time"

	"aurumflow/internal/market"
)

const (
	Unverified         = "UNVERIFIED"
	BrokerMetadataOnly = "BROKER_METADATA_ONLY"
	RuntimeValidated   = "RUNTIME_VALIDATED"
)

// MonetaryInstrumentSpec is venue-agnostic CFD monetary semantics.
type MonetaryInstrumentSpec struct {
	Provider           string
	Epic               string
	Name               string
	InstrumentType     string
	Currency           string
	LotSize            float64
	ContractSize       float64
	ScalingFactor      float64
	PipPosition        int
	TickSize           float64
	MinDealSize        float64
	MaxDealSize        float64
	SizeIncrement      float64
	MoneyPerPriceUnit         float64 // PnL currency per 1.0 price unit per 1.0 size; 0 = unknown
	MoneyPerPriceUnitCurrency string
	ValidationStatus          string
	ValidatedAt               time.Time
	ValidationEvidence        string
	EvidenceVersion           string
	CalibrationSource         string
	CalibrationSamples        int
	CalibrationDealID         string
	Estimator                 string
	PriceMin                  float64
	PriceMax                  float64
	PriceRange                float64
	UPLMin                    float64
	UPLMax                    float64
	UPLRange                  float64
	SlopeMAD                  float64
	DistinctPrices            int
	DistinctUPL               int
	MetadataStatus            string
	PositionClosed            bool
	BrokerPositionsAfter      int
}

func FromMarketDetails(details *market.MarketDetailsResponse) MonetaryInstrumentSpec {
	s := MonetaryInstrumentSpec{Provider: "capital.com", ValidationStatus: Unverified}
	if details == nil {
		return s
	}
	s.Epic = details.Instrument.Epic
	s.Name = details.Instrument.Name
	s.InstrumentType = details.Instrument.Type
	s.Currency = details.Instrument.Currency
	s.LotSize = details.Instrument.LotSize
	s.ContractSize = details.Instrument.ContractSize
	s.ScalingFactor = details.Instrument.ScalingFactor
	s.PipPosition = details.Instrument.PipPosition
	s.TickSize = details.Instrument.TickSize
	s.MinDealSize = details.DealingRules.MinDealSize.Value
	s.MaxDealSize = details.DealingRules.MaxDealSize.Value
	s.SizeIncrement = details.DealingRules.MinSizeIncrement.Value
	if s.DealingComplete() {
		s.ValidationStatus = BrokerMetadataOnly
		if mpu, ok := inferMoneyPerPriceUnit(s); ok {
			s.MoneyPerPriceUnit = mpu
			s.ValidationEvidence = "inferred from lotSize/contractSize/scalingFactor; runtime UPL not yet confirmed"
		}
	}
	return s
}

func inferMoneyPerPriceUnit(s MonetaryInstrumentSpec) (float64, bool) {
	// Capital CFD: when lotSize is present and positive, PnL ≈ size * Δprice * lotSize
	// unless contractSize or scalingFactor indicate otherwise. We only infer when
	// a single positive lotSize is present and other multipliers are 0 or 1.
	if s.LotSize <= 0 {
		return 0, false
	}
	scale := s.ScalingFactor
	if scale == 0 {
		scale = 1
	}
	contract := s.ContractSize
	if contract == 0 {
		contract = 1
	}
	if scale != 1 || contract != 1 {
		return 0, false
	}
	return s.LotSize, true
}

func (s MonetaryInstrumentSpec) DealingComplete() bool {
	return s.Epic != "" && s.MinDealSize > 0 && s.MaxDealSize >= s.MinDealSize && s.SizeIncrement > 0
}

func (s MonetaryInstrumentSpec) StrategyExecutable() bool {
	return s.DealingComplete() && s.MoneyPerPriceUnit > 0 && (s.ValidationStatus == BrokerMetadataOnly || s.ValidationStatus == RuntimeValidated)
}

func (s MonetaryInstrumentSpec) ExpectedPnL(size, openPrice, closePrice float64) (float64, error) {
	if s.MoneyPerPriceUnit <= 0 {
		return 0, fmt.Errorf("monetary spec incomplete: MoneyPerPriceUnit unknown")
	}
	if size <= 0 {
		return 0, fmt.Errorf("invalid size")
	}
	return size * (closePrice - openPrice) * s.MoneyPerPriceUnit, nil
}

func MoneyPerPriceUnit(size float64, s MonetaryInstrumentSpec) (float64, error) {
	if s.MoneyPerPriceUnit <= 0 {
		return 0, fmt.Errorf("MoneyPerPriceUnit unknown")
	}
	return size * s.MoneyPerPriceUnit, nil
}
