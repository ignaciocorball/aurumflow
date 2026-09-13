package journal

import "time"

// RejectReason codes for signal_rejected events.
const (
	RejectNoSignal      = "NO_SIGNAL"
	RejectH1Filter      = "H1_FILTER"
	RejectStateNotReady = "STATE_NOT_READY"
	RejectRiskReject    = "RISK_REJECT"
	RejectSpreadReject  = "SPREAD_REJECT"
	RejectLiveConfirm   = "LIVE_CONFIRM_BLOCK"
	RejectH1Range       = "H1_RANGE_BLOCKED"
	RejectH1RangeCryptoScore = "H1_RANGE_CRYPTO_SCORE"
	RejectH4Range       = "H4_RANGE_BLOCKED"
	RejectH4Filter      = "H4_FILTER"
	RejectMaxTrades     = "MAX_TRADES"
	RejectKillSwitch    = "KILL_SWITCH_ACTIVE"
	RejectMinSizeRisk   = "MIN_SIZE_EXCEEDS_RISK_BUDGET"
	RejectInstrumentSpec = "INSTRUMENT_SPEC_INCOMPLETE"
	RejectExecDisabled  = "EXECUTION_DISABLED"
	RejectUnknownPos    = "UNKNOWN_POSITIONS"
	RejectDryRun        = "ORDER_DRY_RUN"
	RejectMaxSize       = "MAX_SIZE_EXCEEDED"
)

// SetupEvaluated is written when the bot evaluates ComposerInput (every tick in session).
type SetupEvaluated struct {
	Ts                 string   `json:"ts"`
	Event              string   `json:"event"`
	Epic               string   `json:"epic"`
	TfEntry            string   `json:"tf_entry,omitempty"`
	TfBias             string   `json:"tf_bias,omitempty"`
	Session            string   `json:"session"`
	State              string   `json:"state"`
	TrendH4            string   `json:"trend_h4,omitempty"`
	TrendH1            string   `json:"trend_h1"`
	TrendM15           string   `json:"trend_m15"`
	ATR                float64  `json:"atr"`
	RSI                float64  `json:"rsi"`
	ATRBucket          string   `json:"atr_bucket,omitempty"`
	HasLiquidity       bool     `json:"has_liquidity"`
	HasSweep           bool     `json:"has_sweep"`
	AtEquilibrium      bool     `json:"at_equilibrium"`
	Score              int      `json:"score"`
	Threshold          int      `json:"threshold"`
	SignalOK           bool     `json:"signal_ok"`
	SignalDirection    string   `json:"signal_direction,omitempty"`
	RejectReasons      []string `json:"reject_reasons,omitempty"`
}

// SignalRejected is written when a valid signal is rejected by a filter.
type SignalRejected struct {
	Ts          string  `json:"ts"`
	Event       string  `json:"event"`
	Epic        string  `json:"epic"`
	State       string  `json:"state"`
	Session     string  `json:"session"`
	RejectReason string `json:"reject_reason"`
	Direction   string  `json:"direction,omitempty"`
	Entry       float64 `json:"entry,omitempty"`
	Score       int     `json:"score,omitempty"`
}

// SignalGenerated is written when SignalComposerWithZones returns a valid signal.
type SignalGenerated struct {
	Ts               string  `json:"ts"`
	Event            string  `json:"event"`
	Epic             string  `json:"epic"`
	State            string  `json:"state"`
	Session          string  `json:"session"`
	TrendH4          string  `json:"trend_h4,omitempty"`
	TrendH1          string  `json:"trend_h1"`
	Direction        string  `json:"direction"`
	Entry            float64 `json:"entry"`
	SL               float64 `json:"sl"`
	TP               float64 `json:"tp"`
	StopDist         float64 `json:"stop_dist"`
	RR               float64 `json:"rr"`
	Score            int     `json:"score"`
	Confidence       float64 `json:"confidence"`
	ATR              float64 `json:"atr"`
	RSI              float64 `json:"rsi"`
	EquilibriumTouch bool    `json:"equilibrium_touch"`
	SweepType        string  `json:"sweep_type,omitempty"`
	SweepStrength    float64 `json:"sweep_strength,omitempty"`
	LiquidityEvent   bool    `json:"liquidity_event"`
}

// OrderAttempt is written before sending the order to the broker.
type OrderAttempt struct {
	Ts                 string  `json:"ts"`
	Event              string  `json:"event"`
	Epic               string  `json:"epic"`
	Direction          string  `json:"direction"`
	Size               float64 `json:"size"`
	Entry              float64 `json:"entry"`
	SL                 float64 `json:"sl"`
	TP                 float64 `json:"tp"`
	MaxSpread          float64 `json:"max_spread,omitempty"`
	Spread             float64 `json:"spread,omitempty"`
	LiveConfirmRequired bool   `json:"live_confirm_required,omitempty"`
}

// OrderResult is written after broker response (success or error).
type OrderResult struct {
	Ts         string  `json:"ts"`
	Event      string  `json:"event"`
	Epic       string  `json:"epic"`
	Status     string  `json:"status"`
	DealRef    string  `json:"deal_ref,omitempty"`
	DealID     string  `json:"deal_id,omitempty"`
	FillPrice  float64 `json:"fill_price,omitempty"`
	Slippage   float64 `json:"slippage,omitempty"`
	Error      string  `json:"error,omitempty"`
}

// PositionClosed is written when a position closes (live: openCount drops to 0; backtest: each resolved trade).
// Optional segmentación fields (TrendH1, Session, ATRBucket, SweepType, EntryTS, ExitTS) support heatmap analysis.
type PositionClosed struct {
	Ts             string  `json:"ts"`
	Event          string  `json:"event"`
	Epic           string  `json:"epic"`
	DealID         string  `json:"deal_id,omitempty"`
	Direction      string  `json:"direction"`
	Entry          float64 `json:"entry"`
	SL             float64 `json:"sl"`
	TP             float64 `json:"tp"`
	ExitPrice      float64 `json:"exit_price"`
	ExitReason     string  `json:"exit_reason"` // "TP", "SL", "MANUAL", "UNKNOWN"
	PnlR           float64 `json:"pnl_r,omitempty"`
	PnlMoney       float64 `json:"pnl_money,omitempty"`
	MaxFavorable   float64 `json:"max_favorable,omitempty"`
	MaxAdverse     float64 `json:"max_adverse,omitempty"`
	MinutesInTrade int     `json:"minutes_in_trade,omitempty"`
	// Segmentación (backtest / analysis)
	TrendH4   string `json:"trend_h4,omitempty"`
	TrendH1   string `json:"trend_h1,omitempty"`
	Session   string `json:"session,omitempty"`
	ATRBucket string `json:"atr_bucket,omitempty"`
	SweepType string `json:"sweep_type,omitempty"`
	EntryTS   string `json:"entry_ts,omitempty"`
	ExitTS    string `json:"exit_ts,omitempty"`
}

// Event types for JSON "event" field.
const (
	EventSetupEvaluated  = "setup_evaluated"
	EventSignalRejected  = "signal_rejected"
	EventSignalGenerated = "signal_generated"
	EventOrderAttempt    = "order_attempt"
	EventOrderResult     = "order_result"
	EventPositionClosed  = "position_closed"
	EventOrderIntent     = "order_intent"
	EventOrderSkipped    = "order_skipped"
	EventOrderDryRun     = "order_dry_run"
	EventOrderSubmitted  = "order_submitted"
	EventOrderConfirmed  = "order_confirmed"
	EventOrderRejected   = "order_rejected"
	EventPositionOpen    = "position_open"
	EventPositionUpdateRequested = "position_update_requested"
	EventPositionUpdated = "position_updated"
	EventPositionCloseRequested = "position_close_requested"
	EventPositionCloseRejected = "position_close_rejected"
)

func ts() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// Lifecycle is a chronological trade/order event for reconstruction.
type Lifecycle struct {
	Ts            string  `json:"ts"`
	Event         string  `json:"event"`
	Epic          string  `json:"epic,omitempty"`
	SignalID      string  `json:"signal_id,omitempty"`
	State         string  `json:"state,omitempty"`
	ExecutionMode string  `json:"execution_mode,omitempty"`
	Environment   string  `json:"environment,omitempty"`
	DealRef       string  `json:"deal_ref,omitempty"`
	DealID        string  `json:"deal_id,omitempty"`
	Direction     string  `json:"direction,omitempty"`
	Size          float64 `json:"size,omitempty"`
	ConfirmedSize float64 `json:"confirmed_size,omitempty"`
	Entry         float64 `json:"entry,omitempty"`
	SL            float64 `json:"sl,omitempty"`
	TP            float64 `json:"tp,omitempty"`
	FillPrice     float64 `json:"fill_price,omitempty"`
	ClosePrice    float64 `json:"close_price,omitempty"`
	CloseReason   string  `json:"close_reason,omitempty"`
	Status        string  `json:"status,omitempty"`
	Error         string  `json:"error,omitempty"`
	Reason        string  `json:"reason,omitempty"`
	Score         int     `json:"score,omitempty"`
}
