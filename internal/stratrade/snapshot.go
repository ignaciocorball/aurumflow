package stratrade

import "time"

type PreSignalSnapshot struct {
	SignalID                 string    `json:"signal_id"`
	StrategyTradeID          string    `json:"strategy_trade_id"`
	Timestamp                time.Time `json:"timestamp"`
	GitCommit                string    `json:"git_commit"`
	StrategyVersion          string    `json:"strategy_version"`
	Origin                   string    `json:"origin"`
	Instrument               string    `json:"instrument"`
	Direction                string    `json:"direction"`
	Bid                      float64   `json:"bid"`
	Ask                      float64   `json:"ask"`
	Mid                      float64   `json:"mid"`
	Spread                   float64   `json:"spread"`
	MarketStatus             string    `json:"market_status"`
	M5State                  string    `json:"m5_state"`
	H1State                  string    `json:"h1_state"`
	H4State                  string    `json:"h4_state"`
	LegacyScore              int       `json:"legacy_score"`
	LegacyEvidence           string    `json:"legacy_evidence"`
	ATR                      float64   `json:"atr"`
	RSI                      float64   `json:"rsi"`
	Session                  string    `json:"session"`
	StructureState           string    `json:"structure_state"`
	LiquidityState           string    `json:"liquidity_state"`
	IntentState              string    `json:"intent_state"`
	EquilibriumState         string    `json:"equilibrium_state"`
	EntryCandidate           float64   `json:"entry_candidate"`
	StopLoss                 float64   `json:"stop_loss"`
	TakeProfit               float64   `json:"take_profit"`
	StopDistance             float64   `json:"stop_distance"`
	PositionSize             float64   `json:"position_size"`
	MoneyPerPriceUnit        float64   `json:"money_per_price_unit"`
	ExpectedAccountRisk      float64   `json:"expected_account_currency_risk"`
	AccountBalance           float64   `json:"account_balance"`
	DailyPnL                 float64   `json:"daily_pnl"`
	DailyDrawdown            float64   `json:"daily_drawdown"`
	OpenPositionsBefore      int       `json:"open_positions_before"`
	KillSwitch               bool      `json:"kill_switch"`
	BrokerConnection         string    `json:"broker_connection_state"`
	JournalState             string    `json:"journal_state"`
	Immutable                bool      `json:"immutable"`
	WorldHash                string    `json:"world_hash,omitempty"`
	Attention                float64   `json:"attention,omitempty"`
	Coverage                 string    `json:"coverage,omitempty"`
	Salience                 float64   `json:"salience,omitempty"`
	ObservationalNote        string    `json:"observational_note,omitempty"`
	PilotPolicy              string    `json:"pilot_policy,omitempty"`
	PilotRiskLimit           float64   `json:"pilot_risk_limit,omitempty"`
	PortfolioRiskBefore      float64   `json:"portfolio_risk_before,omitempty"`
	StrategyHash             string    `json:"strategy_hash,omitempty"`
	HistoryReady             bool      `json:"history_ready,omitempty"`
}

type IntelligenceContext struct {
	Label              string  `json:"label"`
	WorldHash          string  `json:"world_hash"`
	Liquidity          string  `json:"liquidity"`
	Risk               string  `json:"risk"`
	USD                string  `json:"usd"`
	Rates              string  `json:"rates"`
	Attention          float64 `json:"attention"`
	Coverage           string  `json:"coverage"`
	Tier               string  `json:"tier"`
	CFTCContext        string  `json:"cftc_context"`
	GoldVsSilver       string  `json:"gold_vs_silver"`
	GoldVsEquities     string  `json:"gold_vs_equities"`
	RelativeStrength   string  `json:"live_relative_strength"`
	DataQuality        string  `json:"data_quality"`
	ObservationalOnly  bool    `json:"observational_only"`
}

func NewObservational(ctx IntelligenceContext) IntelligenceContext {
	ctx.Label = LabelObservational
	ctx.ObservationalOnly = true
	return ctx
}

func FreezePreSignal(s PreSignalSnapshot) PreSignalSnapshot {
	s.Immutable = true
	if s.Origin == "" {
		s.Origin = OriginLegacy
	}
	return s
}
