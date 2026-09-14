package mirror

import (
	"errors"
	"strings"

	"aurumflow/internal/research"
)

const (
	SourceNormalized = "LEGACY_NORMALIZED_V0"
	SourceAttention  = "ATTENTION_SCORE_V1"
	SourceSalience   = "MARKET_SALIENCE_V0"
	OriginMirror     = "DEMO_MIRROR"
)

var (
	ErrFake         = errors.New("fake signal rejected")
	ErrResearchOnly = errors.New("Attention/Salience cannot create orders")
	ErrSource       = errors.New("order source is not the frozen validated strategy")
	ErrNoSignal     = errors.New("no genuine strategy signal")
	ErrDuplicate    = errors.New("one signal may create at most one OPEN")
	ErrReconcile    = errors.New("broker/local reconciliation failed")
)

type Intent struct {
	Source string
	Signal research.SignalRow
	Fake   bool
}

func AllowOrder(in Intent) error {
	if in.Fake {
		return ErrFake
	}
	src := strings.ToUpper(strings.TrimSpace(in.Source))
	switch src {
	case SourceAttention, "ATTENTION", SourceSalience, "SALIENCE":
		return ErrResearchOnly
	case SourceNormalized, "LEGACY":
	default:
		return ErrSource
	}
	if in.Signal.Direction == 0 {
		return ErrNoSignal
	}
	return nil
}

func OneSignalOneOrder(seen map[string]bool, signalID string) error {
	if signalID == "" {
		return ErrNoSignal
	}
	if seen[signalID] {
		return ErrDuplicate
	}
	return nil
}

func Reconcile(local, broker, pending int) error {
	if local != broker || pending != 0 {
		return ErrReconcile
	}
	return nil
}

func HaltOnMismatch(local, broker, pending int) bool {
	return Reconcile(local, broker, pending) != nil
}

func ProtectionOK(wantEpic, gotEpic, wantDir, gotDir string, wantSize, gotSize, wantSL, gotSL, wantTP, gotTP, tol float64) bool {
	if !strings.EqualFold(strings.TrimSpace(wantEpic), strings.TrimSpace(gotEpic)) {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(wantDir), strings.TrimSpace(gotDir)) {
		return false
	}
	if mathAbs(wantSize-gotSize) > 1e-6 {
		return false
	}
	if tol <= 0 {
		tol = 0.05
	}
	if wantSL != 0 && (gotSL == 0 || mathAbs(wantSL-gotSL) > tol) {
		return false
	}
	if wantTP != 0 && (gotTP == 0 || mathAbs(wantTP-gotTP) > tol) {
		return false
	}
	return true
}

func EntrySlippage(direction string, decisionBid, decisionAsk, brokerEntry float64) float64 {
	if strings.EqualFold(direction, "SELL") {
		return decisionBid - brokerEntry
	}
	return brokerEntry - decisionAsk
}

func mathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
