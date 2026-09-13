package notifications

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

// Event type constants (report taxonomy).
const (
	TypeHeartbeat        = "HEARTBEAT"
	TypeBotStart         = "BOT_START"
	TypeBotStop          = "BOT_STOP"
	TypeDataStale        = "DATA_STALE"
	TypeAPIError         = "API_ERROR"
	TypeRegimeChange     = "REGIME_CHANGE"
	TypeSweepConfirmed   = "SWEEP_CONFIRMED"
	TypeSignalGenerated  = "SIGNAL_GENERATED"
	TypeSignalRejected   = "SIGNAL_REJECTED"
	TypeNoSignal         = "NO_SIGNAL"
	TypeOrderSent        = "ORDER_SENT"
	TypeOrderFilled      = "ORDER_FILLED"
	TypeOrderRejected    = "ORDER_REJECTED"
	TypePositionOpened   = "POSITION_OPENED"
	TypePositionModified = "POSITION_MODIFIED"
	TypePositionClosed   = "POSITION_CLOSED"
	TypePnlMilestone     = "PNL_MILESTONE"
	TypeDailyDDUpdate    = "DAILY_DD_UPDATE"
	TypeTradingHalted    = "TRADING_HALTED"
	TypeMaxTradesReached = "MAX_TRADES_REACHED"
	TypeDroppedSummary   = "DROPPED_SUMMARY"
)

// Category constants.
const (
	CategoryHealth    = "health"
	CategoryMarket    = "market"
	CategorySignals   = "signals"
	CategoryExecution = "execution"
	CategoryRisk      = "risk"
	CategorySummary   = "summary"
)

// Severity constants (map to Pushover priority).
const (
	SeverityInfo     = "INFO"
	SeverityImportant = "IMPORTANT"
	SeverityCritical  = "CRITICAL"
)

// Standardized reject reason codes for dedup/aggregation (align with journal).
const (
	RejH1BiasRange       = "REJ_H1_BIAS_RANGE"
	RejH4BiasRange       = "REJ_H4_BIAS_RANGE"
	RejH1RangeBlocked    = "REJ_H1_RANGE_BLOCKED"
	RejH1RangeCryptoScore = "REJ_H1_RANGE_CRYPTO_SCORE"
	RejH4RangeBlocked    = "REJ_H4_RANGE_BLOCKED"
	RejH1Filter          = "REJ_H1_FILTER"
	RejH4Filter          = "REJ_H4_FILTER"
	RejSessionBlocked    = "REJ_SESSION_BLOCKED"
	RejScoreBelowThreshold = "REJ_SCORE_BELOW_THRESHOLD"
	RejScoreAfterContext = "REJ_SCORE_AFTER_CONTEXT"
	RejSpreadTooHigh     = "REJ_SPREAD_TOO_HIGH"
	RejDataFrozen        = "REJ_DATA_FROZEN"
	RejRiskDDLimit       = "REJ_RISK_DD_LIMIT"
	RejMaxTrades         = "REJ_MAX_TRADES"
	RejEntryDelayActive  = "REJ_ENTRY_DELAY_ACTIVE"
	RejM5TimingReject    = "REJ_M5_TIMING_REJECT"
	RejLiveConfirm       = "REJ_LIVE_CONFIRM"
	RejStateNotReady     = "REJ_STATE_NOT_READY"
	RejNoSignal          = "REJ_NO_SIGNAL"
	RejRiskReject        = "REJ_RISK_REJECT"
	RejUnknown           = "REJ_UNKNOWN"
)

// NotifEvent is the standard event structure for the notification pipeline.
type NotifEvent struct {
	Type       string
	Category   string
	Severity   string
	Instrument string
	Session    string
	Timestamp  time.Time
	Payload    map[string]any
	DedupKey   string // computed by policy
}

// NormalizeReason normalizes a reject reason string: trim, uppercase, replace spaces with underscore.
func NormalizeReason(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.ToUpper(s)
	s = strings.ReplaceAll(s, " ", "_")
	return s
}

// ReasonToCode maps journal/free-form reject reasons to standard codes (100% deterministic).
// Unknown reasons return "REJ_UNKNOWN:<normalized_raw>". Cap unknown per window in store.
var reasonToCode = map[string]string{
	"H1_FILTER":           RejH1Filter,
	"H1_RANGE_BLOCKED":    RejH1RangeBlocked,
	"H1_RANGE_CRYPTO_SCORE": RejH1RangeCryptoScore,
	"H4_FILTER":           RejH4Filter,
	"H4_RANGE_BLOCKED":    RejH4RangeBlocked,
	"NO_SIGNAL":           RejNoSignal,
	"STATE_NOT_READY":     RejStateNotReady,
	"RISK_REJECT":         RejRiskReject,
	"SPREAD_REJECT":       RejSpreadTooHigh,
	"LIVE_CONFIRM_BLOCK":  RejLiveConfirm,
	"MAX_TRADES":          RejMaxTrades,
	"SCORE_AFTER_CONTEXT": RejScoreAfterContext,
	"M5_TIMING_REJECT":    RejM5TimingReject,
	"H1_BIAS_RANGE":       RejH1BiasRange,
	"H4_BIAS_RANGE":       RejH4BiasRange,
	"SESSION_BLOCKED":     RejSessionBlocked,
	"SCORE_BELOW_THRESHOLD": RejScoreBelowThreshold,
	"DATA_FROZEN":         RejDataFrozen,
	"RISK_DD_LIMIT":       RejRiskDDLimit,
	"ENTRY_DELAY_ACTIVE":  RejEntryDelayActive,
	"KILL_SWITCH_ACTIVE":  RejRiskReject,
	"MIN_SIZE_EXCEEDS_RISK_BUDGET": RejRiskReject,
	"INSTRUMENT_SPEC_INCOMPLETE": RejRiskReject,
	"EXECUTION_DISABLED": RejLiveConfirm,
	"ORDER_DRY_RUN":      RejLiveConfirm,
	"MAX_SIZE_EXCEEDED":  RejRiskReject,
}

// ReasonToCode maps raw reject reason to standard code. Unknown -> "REJ_UNKNOWN:<normalized>".
func ReasonToCode(raw string) string {
	n := NormalizeReason(raw)
	if code, ok := reasonToCode[n]; ok {
		return code
	}
	// Handle "score 4 < threshold 6" style: map to score below
	if strings.Contains(n, "THRESHOLD") || strings.Contains(n, "SCORE") {
		return RejScoreBelowThreshold
	}
	if n == "" {
		return RejNoSignal
	}
	return RejUnknown + ":" + n
}

// DedupKeyHash returns a short hash for DedupKey (Type + Instrument + reason/session/timeframe).
func DedupKeyHash(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}
