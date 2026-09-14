package money

import (
	"math"
	"sort"
	"strings"
	"time"
)

const (
	FailInsufficientPriceExcursion = "INSUFFICIENT_PRICE_EXCURSION"
	FailInsufficientUPLResolution  = "INSUFFICIENT_UPL_RESOLUTION"
	FailInsufficientSamples        = "INSUFFICIENT_SAMPLES"
	FailSlopeUnstable              = "SLOPE_UNSTABLE"
	FailWrongSlopeSign             = "WRONG_SLOPE_SIGN"
	FailMonetaryConflict           = "MONETARY_SEMANTICS_CONFLICT"
	FailBrokerUncertain            = "BROKER_STATE_UNCERTAIN"
	FailPositionReconcile          = "POSITION_RECONCILIATION_FAILED"

	SourceRuntimeUPLSlope = "RUNTIME_UPL_SLOPE"
	EstimatorTheilSen     = "THEIL_SEN"
	MetadataRuntimeAgree  = "METADATA_RUNTIME_AGREE"
	RuntimeEvidenceVer    = "v2-runtime-upl-slope"
)

// Operational estimator gates. Chosen for monetary-contract evidence, not returns.
const (
	MinCalibSamples       = 8
	MinDistinctPrices     = 5
	MinDistinctUPL        = 3
	MinExpectedSignFrac   = 0.70
	MaxRelativeMAD        = 0.50
	MetadataConflictRel   = 0.50
	DefaultSampleInterval = 750 * time.Millisecond
	DefaultCalibMin       = 60 * time.Second
	DefaultCalibMax       = 150 * time.Second
	MaxCalibCanaries      = 2
)

type CalibrationSample struct {
	Timestamp         time.Time
	DealID            string
	Direction         string
	Size              float64
	Bid               float64
	Ask               float64
	CloseablePrice    float64
	BrokerUPL         float64
	PositionOpenLevel float64
	QuoteAge          time.Duration
	PositionAge       time.Duration
	AccountCurrency   string
	EventTime         time.Time
	ReceiveTime       time.Time
}

type CalibrationConfig struct {
	TickSize          float64
	Spread            float64
	MetadataMPU       float64
	MinSamples        int
	MinDistinctPrices int
	MinDistinctUPL    int
	MinExpectedSign   float64
	MaxRelMAD         float64
	MetadataRelTol    float64
}

func DefaultCalibrationConfig() CalibrationConfig {
	return CalibrationConfig{
		MinSamples:        MinCalibSamples,
		MinDistinctPrices: MinDistinctPrices,
		MinDistinctUPL:    MinDistinctUPL,
		MinExpectedSign:   MinExpectedSignFrac,
		MaxRelMAD:         MaxRelativeMAD,
		MetadataRelTol:    MetadataConflictRel,
	}
}

type CalibrationResult struct {
	OK                bool
	Failure           string
	MoneyPerPriceUnit float64
	Currency          string
	Source            string
	Estimator         string
	Samples           int
	DistinctPrices    int
	DistinctUPL       int
	PriceMin          float64
	PriceMax          float64
	PriceRange        float64
	UPLMin            float64
	UPLMax            float64
	UPLRange          float64
	ValidPairSlopes   int
	SlopeMedian       float64
	SlopeP25          float64
	SlopeP75          float64
	SlopeMAD          float64
	ExpectedSignPct   float64
	MetadataMPU       float64
	MetadataStatus    string
	EvidenceDealID    string
	Direction         string
	Size              float64
}

type CloseGuard struct {
	Close  func() error
	closed bool
	err    error
}

func (g *CloseGuard) Ensure() error {
	if g == nil || g.Close == nil || g.closed {
		return g.errOf()
	}
	g.closed = true
	g.err = g.Close()
	return g.err
}

func (g *CloseGuard) errOf() error {
	if g == nil {
		return nil
	}
	return g.err
}

func CloseablePrice(direction string, bid, ask float64) float64 {
	if strings.EqualFold(direction, "SELL") {
		return ask
	}
	return bid
}

func DirectionSign(direction string) float64 {
	if strings.EqualFold(direction, "SELL") {
		return -1
	}
	return 1
}

func EstimateTheilSen(samples []CalibrationSample, cfg CalibrationConfig) CalibrationResult {
	if cfg.MinSamples == 0 {
		cfg = DefaultCalibrationConfig()
	}
	res := CalibrationResult{Source: SourceRuntimeUPLSlope, Estimator: EstimatorTheilSen, MetadataMPU: cfg.MetadataMPU}
	valid := alignedSamples(samples)
	res.Samples = len(valid)
	if len(valid) == 0 {
		res.Failure = FailInsufficientSamples
		return res
	}
	res.EvidenceDealID = valid[0].DealID
	res.Direction = valid[0].Direction
	res.Size = valid[0].Size
	res.Currency = valid[0].AccountCurrency

	prices, upls := values(valid)
	res.DistinctPrices = distinctCount(prices)
	res.DistinctUPL = distinctCount(upls)
	res.PriceMin, res.PriceMax, res.PriceRange = minMaxRange(prices)
	res.UPLMin, res.UPLMax, res.UPLRange = minMaxRange(upls)

	if res.Samples < cfg.MinSamples {
		res.Failure = FailInsufficientSamples
		return res
	}
	need := minExcursion(cfg, prices)
	if res.PriceRange < need {
		res.Failure = FailInsufficientPriceExcursion
		return res
	}
	if res.DistinctPrices < cfg.MinDistinctPrices {
		res.Failure = FailInsufficientSamples
		return res
	}
	if res.DistinctUPL < cfg.MinDistinctUPL || res.UPLRange == 0 {
		res.Failure = FailInsufficientUPLResolution
		return res
	}

	minPair := cfg.TickSize
	if minPair <= 0 {
		minPair = 1e-9
	}
	dir := DirectionSign(valid[0].Direction)
	size := valid[0].Size
	if size <= 0 {
		res.Failure = FailBrokerUncertain
		return res
	}
	var slopes []float64
	for i := 0; i < len(valid); i++ {
		for j := i + 1; j < len(valid); j++ {
			dp := valid[j].CloseablePrice - valid[i].CloseablePrice
			if math.Abs(dp) < minPair {
				continue
			}
			raw := (valid[j].BrokerUPL - valid[i].BrokerUPL) / dp
			norm := dir * raw / size
			if math.IsNaN(norm) || math.IsInf(norm, 0) {
				continue
			}
			slopes = append(slopes, norm)
		}
	}
	res.ValidPairSlopes = len(slopes)
	if len(slopes) < 3 {
		res.Failure = FailInsufficientSamples
		return res
	}
	res.SlopeMedian = percentile(slopes, 0.50)
	res.SlopeP25 = percentile(slopes, 0.25)
	res.SlopeP75 = percentile(slopes, 0.75)
	res.SlopeMAD = mad(slopes, res.SlopeMedian)
	pos := 0
	for _, s := range slopes {
		if s > 0 {
			pos++
		}
	}
	res.ExpectedSignPct = float64(pos) / float64(len(slopes))
	res.MoneyPerPriceUnit = res.SlopeMedian

	if res.SlopeMedian <= 0 {
		res.Failure = FailWrongSlopeSign
		return res
	}
	if res.ExpectedSignPct < cfg.MinExpectedSign {
		res.Failure = FailSlopeUnstable
		return res
	}
	if res.SlopeMedian > 0 && res.SlopeMAD/res.SlopeMedian > cfg.MaxRelMAD {
		res.Failure = FailSlopeUnstable
		return res
	}
	if cfg.MetadataMPU > 0 {
		rel := math.Abs(res.MoneyPerPriceUnit-cfg.MetadataMPU) / math.Max(res.MoneyPerPriceUnit, cfg.MetadataMPU)
		if rel > cfg.MetadataRelTol {
			res.Failure = FailMonetaryConflict
			res.MetadataStatus = FailMonetaryConflict
			return res
		}
		res.MetadataStatus = MetadataRuntimeAgree
	}
	res.OK = true
	return res
}

func ApplyRuntime(spec MonetaryInstrumentSpec, res CalibrationResult) MonetaryInstrumentSpec {
	if !res.OK || res.MoneyPerPriceUnit <= 0 || res.Failure != "" {
		if spec.ValidationStatus == RuntimeValidated {
			spec.ValidationStatus = BrokerMetadataOnly
		}
		if res.Failure != "" {
			spec.ValidationEvidence = res.Failure
		}
		return spec
	}
	spec.MoneyPerPriceUnit = res.MoneyPerPriceUnit
	spec.MoneyPerPriceUnitCurrency = res.Currency
	spec.ValidationStatus = RuntimeValidated
	spec.ValidationEvidence = SourceRuntimeUPLSlope + " " + EstimatorTheilSen
	spec.EvidenceVersion = RuntimeEvidenceVer
	spec.CalibrationSource = SourceRuntimeUPLSlope
	spec.CalibrationSamples = res.Samples
	spec.CalibrationDealID = res.EvidenceDealID
	spec.ValidatedAt = time.Now().UTC()
	spec.Estimator = res.Estimator
	spec.PriceMin = res.PriceMin
	spec.PriceMax = res.PriceMax
	spec.PriceRange = res.PriceRange
	spec.UPLMin = res.UPLMin
	spec.UPLMax = res.UPLMax
	spec.UPLRange = res.UPLRange
	spec.SlopeMAD = res.SlopeMAD
	spec.DistinctPrices = res.DistinctPrices
	spec.DistinctUPL = res.DistinctUPL
	spec.MetadataStatus = res.MetadataStatus
	return spec
}

func alignedSamples(in []CalibrationSample) []CalibrationSample {
	var out []CalibrationSample
	var deal, dir string
	var size float64
	for _, s := range in {
		if s.DealID == "" || s.Size <= 0 || s.CloseablePrice <= 0 {
			continue
		}
		if math.IsNaN(s.BrokerUPL) || math.IsInf(s.BrokerUPL, 0) {
			continue
		}
		if deal == "" {
			deal, dir, size = s.DealID, s.Direction, s.Size
		}
		if s.DealID != deal || !strings.EqualFold(s.Direction, dir) || s.Size != size {
			continue
		}
		out = append(out, s)
	}
	return out
}

func values(xs []CalibrationSample) (prices, upls []float64) {
	prices = make([]float64, len(xs))
	upls = make([]float64, len(xs))
	for i, s := range xs {
		prices[i] = s.CloseablePrice
		upls[i] = s.BrokerUPL
	}
	return prices, upls
}

func distinctCount(xs []float64) int {
	seen := map[float64]struct{}{}
	for _, v := range xs {
		seen[v] = struct{}{}
	}
	return len(seen)
}

func minMaxRange(xs []float64) (min, max, rng float64) {
	if len(xs) == 0 {
		return 0, 0, 0
	}
	min, max = xs[0], xs[0]
	for _, v := range xs[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max, max - min
}

func minExcursion(cfg CalibrationConfig, prices []float64) float64 {
	need := 0.0
	if cfg.TickSize > 0 {
		need = 2 * cfg.TickSize
	}
	if cfg.Spread > 0 {
		if s := 0.25 * cfg.Spread; s > need {
			need = s
		}
	}
	if need > 0 {
		return need
	}
	step := smallestPositiveStep(prices)
	if step > 0 {
		return 4 * step
	}
	return 0
}

func smallestPositiveStep(xs []float64) float64 {
	cp := append([]float64{}, xs...)
	sort.Float64s(cp)
	min := 0.0
	for i := 1; i < len(cp); i++ {
		d := cp[i] - cp[i-1]
		if d <= 0 {
			continue
		}
		if min == 0 || d < min {
			min = d
		}
	}
	return min
}

func percentile(xs []float64, p float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	cp := append([]float64{}, xs...)
	sort.Float64s(cp)
	if p <= 0 {
		return cp[0]
	}
	if p >= 1 {
		return cp[len(cp)-1]
	}
	idx := p * float64(len(cp)-1)
	lo := int(math.Floor(idx))
	hi := int(math.Ceil(idx))
	if lo == hi {
		return cp[lo]
	}
	w := idx - float64(lo)
	return cp[lo]*(1-w) + cp[hi]*w
}

func mad(xs []float64, med float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	dev := make([]float64, len(xs))
	for i, v := range xs {
		dev[i] = math.Abs(v - med)
	}
	return percentile(dev, 0.50)
}
