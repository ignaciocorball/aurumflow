package terminal

import "time"

type Snapshot struct {
	GeneratedAt string          `json:"generated_at"`
	FixtureMode bool            `json:"fixture_mode"`
	Environment Environment     `json:"environment"`
	World       WorldView       `json:"world"`
	Markets     []Market        `json:"markets"`
	Portfolio   Portfolio       `json:"portfolio"`
	Execution   ExecutionView   `json:"execution"`
	Research    ResearchView    `json:"research"`
	News        []NewsCluster   `json:"news"`
	Events      []EconEvent     `json:"events"`
	Health      []HealthItem    `json:"health"`
	Narratives  Narratives      `json:"narratives"`
	Activity    []ActivityItem  `json:"activity"`
	Micro       MicroView       `json:"micro"`
	Regions     []RegionView    `json:"regions"`
	Flows       []FlowView      `json:"flows"`
}

type Environment struct {
	Name         string `json:"name"`
	Demo         bool   `json:"demo"`
	LivePossible string `json:"live_possible"`
	LivePossibleDisplay string `json:"live_possible_display"`
	AccountMasked string `json:"account_masked"`
	BrokerStatus string `json:"broker_status"`
	WorldValid   string `json:"world_valid"`
	WorldValidDisplay string `json:"world_valid_display"`
	DataHealth   string `json:"data_health"`
	UTC          string `json:"utc"`
}

type WorldView struct {
	AsOf         string  `json:"as_of"`
	Valid        string  `json:"valid"`
	ValidDisplay string  `json:"valid_display"`
	Hash         string  `json:"hash"`
	Session      string  `json:"session"`
	Regime       string  `json:"regime"`
	RegimeDisplay string `json:"regime_display"`
	Dislocation  string  `json:"dislocation"`
	DislocationDisplay string `json:"dislocation_display"`
	Confidence   float64 `json:"confidence"`
	Liquidity    string  `json:"liquidity"`
	USD          string  `json:"usd"`
}

type Market struct {
	ID              string    `json:"id"`
	Label           string    `json:"label"`
	Region          string    `json:"region"`
	RegionLabel     string    `json:"region_label"`
	AssetClass      string    `json:"asset_class"`
	Bid             float64   `json:"bid"`
	Ask             float64   `json:"ask"`
	Mid             float64   `json:"mid"`
	ChangePct       float64   `json:"change_pct"`
	Spark           []float64 `json:"spark"`
	Attention       float64   `json:"attention"`
	Salience        float64   `json:"salience"`
	Coverage        float64   `json:"coverage"`
	Trend           string    `json:"trend"`
	TrendDisplay    string    `json:"trend_display"`
	Regime          string    `json:"regime"`
	Setup           string    `json:"setup"`
	SetupDisplay    string    `json:"setup_display"`
	Session         string    `json:"session"`
	SessionDisplay  string    `json:"session_display"`
	Research        string    `json:"research"`
	ResearchDisplay string    `json:"research_display"`
	DemoStatus      string    `json:"demo_status"`
	DemoDisplay     string    `json:"demo_display"`
	Quality         string    `json:"quality"`
	QualityDisplay  string    `json:"quality_display"`
	MarketStatus    string    `json:"market_status"`
	MarketStatusDisplay string `json:"market_status_display"`
	Eligibility     string    `json:"eligibility"`
	Positioning     string    `json:"positioning"`
	FlowContext     string    `json:"flow_context"`
	SalienceReason  string    `json:"salience_reason"`
	Evidence        []string  `json:"evidence"`
	Pilot           bool      `json:"pilot"`
}

type RegionView struct {
	ID            string   `json:"id"`
	Label         string   `json:"label"`
	Salience      float64  `json:"salience"`
	Attention     float64  `json:"attention"`
	Coverage      float64  `json:"coverage"`
	Quality       string   `json:"quality"`
	Markets       []string `json:"markets"`
	Narrative     string   `json:"narrative"`
	NextEvent     string   `json:"next_event"`
}

type FlowView struct {
	From        string  `json:"from"`
	To          string  `json:"to"`
	Class       string  `json:"class"`
	ClassDisplay string `json:"class_display"`
	Evidence    string  `json:"evidence"`
	Confidence  string  `json:"confidence"`
	Strength    float64 `json:"strength"`
}

type Portfolio struct {
	Equity           float64     `json:"equity"`
	Currency         string      `json:"currency"`
	TodayPnL         float64     `json:"today_pnl"`
	TodayR           float64     `json:"today_r"`
	Unrealized       float64     `json:"unrealized"`
	Realized         float64     `json:"realized"`
	Net              float64     `json:"net"`
	ReturnPct        float64     `json:"return_pct"`
	DrawdownPct      float64     `json:"drawdown_pct"`
	OpenRisk         float64     `json:"open_risk"`
	OpenRiskCap      float64     `json:"open_risk_cap"`
	PilotCap         float64     `json:"pilot_cap"`
	AggregateCap     float64     `json:"aggregate_cap"`
	PreciousRisk     float64     `json:"precious_risk"`
	PreciousCap      float64     `json:"precious_cap"`
	USEquityRisk     float64     `json:"us_equity_risk"`
	USEquityCap      float64     `json:"us_equity_cap"`
	Scope            string      `json:"scope"`
	PilotsArmed      int         `json:"pilots_armed"`
	Calendar         []DayPnL    `json:"calendar"`
	WeeklyTotal      float64     `json:"weekly_total"`
	MonthlyTotal     float64     `json:"monthly_total"`
	AllTime          float64     `json:"all_time"`
	AllTimeR         float64     `json:"all_time_r"`
	TradeCount       int         `json:"trade_count"`
	EquityCurve      []CurvePt   `json:"equity_curve"`
	HasBrokerTrades  bool        `json:"has_broker_trades"`
}

type DayPnL struct {
	Date       string  `json:"date"`
	Realized   float64 `json:"realized"`
	Unrealized float64 `json:"unrealized"`
	R          float64 `json:"r"`
	Trades     int     `json:"trades"`
	Wins       int     `json:"wins"`
	Losses     int     `json:"losses"`
	Best       float64 `json:"best"`
	Worst      float64 `json:"worst"`
}

type CurvePt struct {
	T string  `json:"t"`
	V float64 `json:"v"`
}

type ExecutionView struct {
	Banner     string     `json:"banner"`
	Positions  []Position `json:"positions"`
	RiskGroups []RiskBar  `json:"risk_groups"`
	Halt       bool       `json:"halt"`
	Trust      []TrustItem `json:"trust"`
}

type Position struct {
	Market          string      `json:"market"`
	Origin          string      `json:"origin"`
	OriginDisplay   string      `json:"origin_display"`
	Kind            string      `json:"kind"` // DEMO | SHADOW | RESEARCH
	Strategy        string      `json:"strategy"`
	Direction       string      `json:"direction"`
	Entry           float64     `json:"entry"`
	Current         float64     `json:"current"`
	SL              float64     `json:"sl"`
	TP              float64     `json:"tp"`
	UPL             float64     `json:"upl"`
	ReturnR         float64     `json:"return_r"`
	PlannedRisk     float64     `json:"planned_risk"`
	ActualRisk      float64     `json:"actual_risk"`
	Account         string      `json:"account"`
	Open            bool        `json:"open"`
	Lifecycle       []LifeStage `json:"lifecycle"`
	LifecycleLabel  string      `json:"lifecycle_label"`
	BrokerReconcile string      `json:"broker_reconcile"`
	Session         string      `json:"session"`
	Waiting         string      `json:"waiting"`
	LastExecution   string      `json:"last_execution"`
	DealRef         string      `json:"deal_ref"`
	Size            float64     `json:"size"`
	Eligible        string      `json:"eligible"`
	Armed           string      `json:"armed"`
}

type LifeStage struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	State  string `json:"state"` // done | active | pending | skip
}

type RiskBar struct {
	ID      string  `json:"id"`
	Label   string  `json:"label"`
	Used    float64 `json:"used"`
	Cap     float64 `json:"cap"`
}

type TrustItem struct {
	Code    string `json:"code"`
	Display string `json:"display"`
	Hint    string `json:"hint"`
	Market  string `json:"market"`
}

type ResearchView struct {
	Rows []ResearchRow `json:"rows"`
}

type ResearchRow struct {
	Market     string  `json:"market"`
	Status     string  `json:"status"`
	StatusDisplay string `json:"status_display"`
	HoldoutN   int     `json:"holdout_n"`
	DiscoveryN int     `json:"discovery_n"`
	Expectancy float64 `json:"expectancy"`
	ProfitFactor float64 `json:"profit_factor"`
	Hit        float64 `json:"hit"`
	MaxDD      float64 `json:"max_dd"`
	Sample     int     `json:"sample"`
	Spec       string  `json:"spec"`
	SpecHash   string  `json:"spec_hash"`
	Note       string  `json:"note"`
	Dataset    string  `json:"dataset"`
}

type NewsCluster struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	Source     string   `json:"source"`
	Sources    []string `json:"sources"`
	URL        string   `json:"url"`
	Published  string   `json:"published"`
	Age        string   `json:"age"`
	Markets    []string `json:"markets"`
	Themes     []string `json:"themes"`
	Relevance  float64  `json:"relevance"`
	Importance string   `json:"importance"`
	Count      int      `json:"count"`
	Region     string   `json:"region"`
}

type EconEvent struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Region        string   `json:"region"`
	At            string   `json:"at"`
	Importance    string   `json:"importance"`
	Source        string   `json:"source"`
	SourceURL     string   `json:"source_url"`
	AssetClasses  []string `json:"asset_classes"`
	Markets       []string `json:"markets"`
	Status        string   `json:"status"`
	Window        string   `json:"window"`
	Countdown     string   `json:"countdown"`
	Actual        string   `json:"actual"`
	Forecast      string   `json:"forecast"`
	Previous      string   `json:"previous"`
}

type HealthItem struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Status     string `json:"status"`
	Display    string `json:"display"`
	LastUpdate string `json:"last_update"`
	Detail     string `json:"detail"`
}

type Narratives struct {
	World     string   `json:"world"`
	Matters   []string `json:"matters"`
	Execution string   `json:"execution"`
	Health    string   `json:"health"`
}

type ActivityItem struct {
	At      string `json:"at"`
	Kind    string `json:"kind"`
	Market  string `json:"market"`
	Text    string `json:"text"`
	Process string `json:"process"`
}

type MicroView struct {
	Price       float64     `json:"price"`
	Microprice  float64     `json:"microprice"`
	Spread      float64     `json:"spread"`
	CVD         float64     `json:"cvd"`
	Pressure    float64     `json:"pressure"`
	Imbalance   float64     `json:"imbalance"`
	Imb1        float64     `json:"imb1"`
	Imb5        float64     `json:"imb5"`
	Imb10       float64     `json:"imb10"`
	BookSynced  bool        `json:"book_synced"`
	BookAgeMs   int64       `json:"book_age_ms"`
	Gaps        int         `json:"gaps"`
	Resyncs     int         `json:"resyncs"`
	Drops       int64       `json:"drops"`
	LatencyP50  float64     `json:"latency_p50"`
	LatencyP95  float64     `json:"latency_p95"`
	LatencyP99  float64     `json:"latency_p99"`
	Provider    string      `json:"provider"`
	Capability  string      `json:"capability"`
	Absorption  string      `json:"absorption"`
	Exhaustion  string      `json:"exhaustion"`
	Integrity   string      `json:"integrity"`
	EventRate   float64     `json:"event_rate"`
	Series      []MicroPt   `json:"series"`
	Bids        []BookLvl   `json:"bids"`
	Asks        []BookLvl   `json:"asks"`
	ResyncReason string     `json:"resync_reason"`
}

type MicroPt struct {
	T        int64   `json:"t"`
	Price    float64 `json:"price"`
	Micro    float64 `json:"microprice"`
	Pressure float64 `json:"pressure"`
	CVD      float64 `json:"cvd"`
}

type BookLvl struct {
	Price float64 `json:"price"`
	Qty   float64 `json:"qty"`
}

type StreamEvent struct {
	Type string      `json:"type"`
	At   string      `json:"at"`
	Data any         `json:"data,omitempty"`
}

func nowUTC() string { return time.Now().UTC().Format(time.RFC3339) }
