package ops

import "time"

type Status struct {
	UptimeSeconds     int64  `json:"uptime_seconds"`
	APIEnvironment    string `json:"api_environment"`
	ExecutionMode     string `json:"execution_mode"`
	Host              string `json:"host"`
	KillSwitch        bool   `json:"kill_switch"`
	OpenPositions     int    `json:"open_positions"`
	DailyPnL          float64 `json:"daily_pnl"`
	DailyDDPct        float64 `json:"daily_dd_pct"`
	ExecutionEpic     string `json:"execution_epic"`
	MarketStatus      string `json:"market_status"`
	RadarMode         string `json:"radar_mode"`
	RadarProvider     string `json:"radar_provider"`
	BookSynced        bool   `json:"book_synced"`
	EventRate         float64 `json:"event_rate"`
	Reconnects        int    `json:"reconnects"`
	Resyncs           int    `json:"resyncs"`
	CVD               float64 `json:"cvd"`
	Imbalance         float64 `json:"imbalance"`
	Absorption        float64 `json:"absorption"`
	Pressure          float64 `json:"pressure"`
	RadarState        string `json:"radar_state"`
	Confidence        float64 `json:"confidence"`
	LastStrategy      string `json:"last_strategy"`
	LastExecution     string `json:"last_execution"`
	LastError         string `json:"last_error"`
	CME_NQ            string `json:"cme_nq"`
	CME_GC            string `json:"cme_gc"`
	NasdaqTotalView   string `json:"nasdaq_totalview"`
	Options           string `json:"options"`
	LastLegacyDir     int     `json:"last_legacy_direction"`
	LastPressure      float64 `json:"last_pressure_score"`
	LastV1Class       string  `json:"last_v1_classification"`
	FlowEfficiency    float64 `json:"flow_efficiency"`
	ImpactFailure     float64 `json:"impact_failure"`
	BookCapability    string  `json:"book_capability"`
	FeedFreshness     string  `json:"feed_freshness"`
	ProspectiveTotal  int     `json:"prospective_signals"`
	ProspectiveExh    int     `json:"prospective_exhaustion"`
	ProspectiveCont   int     `json:"prospective_continuation"`
	ProspectiveNeu    int     `json:"prospective_neutral"`
	Mature15m         int     `json:"mature_15m"`
	Mature1h          int     `json:"mature_1h"`
	BookAgeMs         int64   `json:"book_age_ms"`
	BookGaps          int     `json:"book_gaps"`
	Vel1s             float64 `json:"flow_velocity_1s"`
	Vel5s             float64 `json:"flow_velocity_5s"`
	Vel30s            float64 `json:"flow_velocity_30s"`
	AbsorptionStatus  string  `json:"absorption_status"`
	SupportingRepl    float64 `json:"supporting_replenishment"`
	OpposingDepl      float64 `json:"opposing_depletion"`
	SupportingPers    float64 `json:"supporting_persistence"`
	PrimaryL2         string  `json:"primary_l2_sensor"`
	CapitalEnv        string  `json:"capital_environment"`
	LivePossible      string  `json:"live_possible"`
	DemoBalance       float64 `json:"demo_balance"`
	RadarShadow       string  `json:"radar_shadow"`
	ExhaustionShadow  string  `json:"exhaustion_shadow"`
	AbsorptionShadow  string  `json:"absorption_shadow"`
	GoldBid           float64 `json:"gold_bid"`
	GoldAsk           float64 `json:"gold_ask"`
	GoldSpread        float64 `json:"gold_spread"`
	GoldValidation    string  `json:"gold_validation"`
	GoldSession       string  `json:"gold_session"`
	GoldEntry         float64 `json:"gold_entry"`
	GoldSL            float64 `json:"gold_sl"`
	GoldTP            float64 `json:"gold_tp"`
	GoldUPnL          float64 `json:"gold_upnl"`
	TradesToday       int     `json:"trades_today"`
	BTCPrice          float64 `json:"btc_price"`
	AggBuy            float64 `json:"aggressive_buy_flow"`
	AggSell           float64 `json:"aggressive_sell_flow"`
	DirectionalP      float64 `json:"directional_pressure"`
	L2Provider        string  `json:"l2_provider"`
	L2Instrument      string  `json:"l2_instrument"`
	L2Relation        string  `json:"l2_relation"`
	L2ProxyQuality    string  `json:"l2_proxy_quality"`
	Microprice        float64 `json:"microprice"`
	Imb1              float64 `json:"imbalance_1"`
	Imb5              float64 `json:"imbalance_5"`
	Imb10             float64 `json:"imbalance_10"`
	Imb20             float64 `json:"imbalance_20"`
	BidRepl           float64 `json:"bid_replenishment"`
	AskRepl           float64 `json:"ask_replenishment"`
	BidDepl           float64 `json:"bid_depletion"`
	AskDepl           float64 `json:"ask_depletion"`
	BidPersist        float64 `json:"bid_persistence"`
	AskPersist        float64 `json:"ask_persistence"`
	AbsorptionWhy     string  `json:"absorption_why"`
	AbsorptionEvidence string `json:"absorption_evidence"`
	V1FlowProvider    string  `json:"v1_flow_provider"`
	EventFreshness    string  `json:"event_freshness"`
	BookFreshness     string  `json:"book_freshness"`
	LastUpdate        string  `json:"last_update"`
	Deltas            int64   `json:"depth_deltas"`
	Snapshots         int64   `json:"book_snapshots"`
	Drops             int64   `json:"dropped_events"`
	LatencyP50        float64 `json:"latency_p50_ms"`
	LatencyP95        float64 `json:"latency_p95_ms"`
	LatencyP99        float64 `json:"latency_p99_ms"`
	DiskMB            float64 `json:"collector_disk_mb"`
	Basis             float64 `json:"cross_venue_basis"`
	BasisZ            float64 `json:"basis_z"`
	ExhL2Unavailable  int     `json:"exh_l2_unavailable"`
	ExhNotSupportive  int     `json:"exh_not_supportive"`
	ExhMixed          int     `json:"exh_mixed"`
	ExhSupportive     int     `json:"exh_supportive"`
	ExhStrong         int     `json:"exh_strongly_supportive"`
	ContL2            int     `json:"continuation_with_l2"`
	Alerts            []string `json:"alerts,omitempty"`
	L2Spread          float64  `json:"l2_spread"`
	L2Mid             float64  `json:"l2_mid"`
	L2QuotesOK        bool     `json:"l2_quotes_ok"`
	GoldQuotesOK      bool     `json:"gold_quotes_ok"`
	DemoBalanceOK     bool     `json:"demo_balance_ok"`
	PositionOpen      bool     `json:"position_open"`
	PeakMemMB         float64  `json:"peak_mem_mb"`
	TopBids           []BookLevel `json:"top_bids,omitempty"`
	TopAsks           []BookLevel `json:"top_asks,omitempty"`
	DecisionWhy       string      `json:"decision_why"`
	LastEvent         string      `json:"last_event"`
	PositionsKnown    bool        `json:"positions_known"`
	Trades            int64       `json:"trades"`
	StrategySession   string      `json:"strategy_session"`
	StrategyReady     bool        `json:"strategy_ready"`
	StrategyWaiting   string      `json:"strategy_waiting"`
	NextSession       string      `json:"next_session"`
	NextSessionAt     string      `json:"next_session_at"`
	LastScan          string      `json:"last_scan"`
	LastSignal        string      `json:"last_signal"`
	TradeCount        int         `json:"trade_count"`
	MaxTrades         int         `json:"max_trades"`
	MonetaryStatus    string      `json:"monetary_status"`
	OperationalTrust  string      `json:"operational_trust"`
	ExpectedRisk      float64     `json:"expected_risk"`
	HoldingSeconds    int64       `json:"holding_seconds"`
	ObservationalNote string      `json:"observational_note"`
}

func NewStatus() Status {
	return Status{
		CME_NQ: "NOT_CONNECTED", CME_GC: "NOT_CONNECTED",
		NasdaqTotalView: "NOT_CONNECTED", Options: "NOT_CONNECTED",
		RadarMode: "SHADOW",
		CapitalEnv: "DEMO",
		LivePossible: "IMPOSSIBLE / FAIL-CLOSED",
		RadarShadow: "SHADOW",
		ExhaustionShadow: "SHADOW",
		AbsorptionShadow: "SHADOW",
		V1FlowProvider: "binance_usdm_public",
	}
}

func AgeSeconds(start time.Time) int64 {
	if start.IsZero() {
		return 0
	}
	return int64(time.Since(start).Seconds())
}
