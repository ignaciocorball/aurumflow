package stratrade

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

const (
	OriginLegacy       = "LEGACY_STRATEGY"
	OriginCalibration  = "CALIBRATION"
	LabelObservational = "OBSERVATIONAL_CONTEXT_ONLY"

	EvSignalObserved          = "SIGNAL_OBSERVED"
	EvRiskEvaluated           = "RISK_EVALUATED"
	EvRiskAccepted            = "RISK_ACCEPTED"
	EvRiskRejected            = "RISK_REJECTED"
	EvOrderIntent             = "ORDER_INTENT_CREATED"
	EvBrokerOpenRequest       = "BROKER_OPEN_REQUEST"
	EvBrokerConfirm           = "BROKER_CONFIRMATION_RECEIVED"
	EvPositionResolved        = "POSITION_IDENTITY_RESOLVED"
	EvPositionReconciled      = "POSITION_RECONCILED"
	EvProtectiveVerified      = "PROTECTIVE_LEVELS_VERIFIED"
	EvMonitoring              = "POSITION_MONITORING"
	EvExitCondition           = "EXIT_CONDITION"
	EvBrokerCloseRequest      = "BROKER_CLOSE_REQUEST"
	EvBrokerCloseConfirm      = "BROKER_CLOSE_CONFIRMATION"
	EvPositionClosed          = "POSITION_CLOSED"
	EvFinalReconciliation     = "FINAL_RECONCILIATION"

	TradeWin        = "WIN"
	TradeLoss       = "LOSS"
	TradeBreakeven  = "BREAKEVEN"
	OpsClean        = "CLEAN"
	OpsWarning      = "WARNING"
	OpsFailed       = "FAILED"

	TrustPASS     = "PASS"
	TrustPartial  = "PARTIAL"
	TrustFAIL     = "FAIL"
	TrustUntested = "UNTESTED"
)

func TradeID(signalID, dealRef, dealID, epic string) string {
	epic = strings.ToUpper(strings.TrimSpace(epic))
	raw := strings.Join([]string{epic, signalID, dealRef, dealID}, "|")
	sum := sha256.Sum256([]byte(raw))
	return epic + "-" + hex.EncodeToString(sum[:8])
}

func SignalID(epic string, t time.Time, direction string) string {
	return strings.ToUpper(epic) + "-" + t.UTC().Format("20060102T150405Z") + "-" + strings.ToUpper(direction)
}

func IsStrategyOrigin(origin string) bool {
	return origin == OriginLegacy
}
