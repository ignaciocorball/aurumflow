package worlddomain

import "time"

type Region string
type AssetClass string
type Frequency string
type SensorHealth string
type EvidenceKind string
type DataRelation string
type FlowDir string
type RiskRegime string
type LiquidityClass string
type AttentionTier string
type ExecEligibility string
type SetupState string
type Origin string
type SourceMode string

const (
	RegionGlobal     Region = "GLOBAL"
	RegionUS         Region = "UNITED_STATES"
	RegionEurope     Region = "EUROPE"
	RegionJapan      Region = "JAPAN"
	RegionChinaHK    Region = "CHINA_HONG_KONG"
	RegionAsiaExJP   Region = "ASIA_EX_JAPAN"

	AssetEquities   AssetClass = "EQUITIES"
	AssetFixedInc   AssetClass = "FIXED_INCOME"
	AssetCredit     AssetClass = "CREDIT"
	AssetFX         AssetClass = "FX"
	AssetPrecious   AssetClass = "PRECIOUS_METALS"
	AssetEnergy     AssetClass = "ENERGY"
	AssetIndMetals  AssetClass = "INDUSTRIAL_METALS"
	AssetCrypto     AssetClass = "CRYPTO"

	FreqLive      Frequency = "LIVE"
	FreqIntraday  Frequency = "INTRADAY"
	FreqDaily     Frequency = "DAILY"
	FreqWeekly    Frequency = "WEEKLY"
	FreqMonthly   Frequency = "MONTHLY"
	FreqQuarterly Frequency = "QUARTERLY"
	FreqRelease   Frequency = "RELEASE"

	HealthHealthy     SensorHealth = "HEALTHY"
	HealthStale       SensorHealth = "STALE"
	HealthDegraded    SensorHealth = "DEGRADED"
	HealthUnavailable SensorHealth = "UNAVAILABLE"
	HealthUnknown     SensorHealth = "UNKNOWN"

	KindObserved EvidenceKind = "OBSERVED"
	KindDerived  EvidenceKind = "DERIVED"
	KindInferred EvidenceKind = "INFERRED"

	RelDirect           DataRelation = "DIRECT"
	RelProxy            DataRelation = "PROXY"
	RelCorrelatedProxy  DataRelation = "CORRELATED_PROXY"
	RelDelayed          DataRelation = "DELAYED"
	RelSlowContext      DataRelation = "SLOW_CONTEXT"
	RelExperimental     DataRelation = "EXPERIMENTAL"

	FlowIn    FlowDir = "INFLOW"
	FlowOut   FlowDir = "OUTFLOW"
	FlowMixed FlowDir = "MIXED"
	FlowUnk   FlowDir = "UNKNOWN"

	RiskOn         RiskRegime = "RISK_ON"
	RiskOff        RiskRegime = "RISK_OFF"
	RiskMixed      RiskRegime = "MIXED"
	RiskTransition RiskRegime = "TRANSITION"

	LiqExpanding   LiquidityClass = "EXPANDING"
	LiqNeutral     LiquidityClass = "NEUTRAL"
	LiqContracting LiquidityClass = "CONTRACTING"
	LiqUnknown     LiquidityClass = "UNKNOWN"

	TierA    AttentionTier = "TIER_A"
	TierB    AttentionTier = "TIER_B"
	TierC    AttentionTier = "TIER_C"
	TierIgnore AttentionTier = "IGNORE"

	EligAnalysis     ExecEligibility = "ANALYSIS_ONLY"
	EligNotCalibrated ExecEligibility = "DEMO_NOT_CALIBRATED"
	EligDemo         ExecEligibility = "DEMO_ELIGIBLE"
	EligLiveProhibited ExecEligibility = "LIVE_PROHIBITED"
	EligDiscovered     ExecEligibility = "DEMO_DISCOVERED"
	EligSpecValid      ExecEligibility = "DEMO_SPEC_VALID"
	EligCalibrated     ExecEligibility = "DEMO_CALIBRATED"
	EligBlocked        ExecEligibility = "BLOCKED"

	OriginLive    Origin = "LIVE_OFFICIAL"
	OriginCache   Origin = "CACHE_OFFICIAL"
	OriginFixture Origin = "FIXTURE"

	ModeFixture SourceMode = "FIXTURE"
	ModeLive    SourceMode = "LIVE_OFFICIAL"
	ModeCache   SourceMode = "CACHE_OFFICIAL"

	SetupNone      SetupState = "NONE"
	SetupPotential SetupState = "POTENTIAL"
	SetupBlocked   SetupState = "BLOCKED"
)

type ContextObservation struct {
	Source      string
	Region      Region
	AssetClass  AssetClass
	Market      string
	Metric      string
	Value       float64
	Unit        string
	Present     bool
	ObservedAt  time.Time
	PublishedAt time.Time
	AvailableAt time.Time
	RetrievedAt time.Time
	Period      string
	Frequency   Frequency
	Quality     SensorHealth
	Relation    DataRelation
	Kind        EvidenceKind
	SourceURL   string
	Origin      Origin
	DatasetID   string
	SeriesID    string
	ParserVer   string
	RawHash     string
}

func (o ContextObservation) UsableAt(t time.Time) bool {
	if !o.Present {
		return false
	}
	avail := o.AvailableAt
	if avail.IsZero() {
		avail = o.PublishedAt
	}
	if avail.IsZero() {
		avail = o.ObservedAt
	}
	if o.PublishedAt.After(t.UTC()) {
		return false
	}
	return !avail.After(t.UTC())
}

func ProductionOrigin(o Origin) bool {
	return o == OriginLive || o == OriginCache
}

type Provenance struct {
	Source    string
	URL       string
	AsOf      time.Time
	Available time.Time
	Freshness Frequency
	Health    SensorHealth
	Kind      EvidenceKind
}

type CapitalFlowVector struct {
	Region     Region
	AssetClass AssetClass
	Direction  FlowDir
	Strength   float64
	Freshness  Frequency
	Health     SensorHealth
	Evidence   []string
}

type SensorEvidence struct {
	Source  string
	Metric  string
	Text    string
	Kind    EvidenceKind
	Health  SensorHealth
	AsOf    time.Time
}

func MissingLabel() string { return "—" }

func HealthFromAge(avail time.Time, now time.Time, staleAfter, unavailAfter time.Duration) SensorHealth {
	if avail.IsZero() {
		return HealthUnknown
	}
	age := now.UTC().Sub(avail.UTC())
	if age < 0 {
		age = 0
	}
	switch {
	case age > unavailAfter:
		return HealthUnavailable
	case age > staleAfter:
		return HealthStale
	default:
		return HealthHealthy
	}
}
