package stratrade

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var RequiredLifecycle = []string{
	EvSignalObserved,
	EvRiskEvaluated,
	EvRiskAccepted,
	EvOrderIntent,
	EvBrokerOpenRequest,
	EvBrokerConfirm,
	EvPositionResolved,
	EvPositionReconciled,
	EvProtectiveVerified,
	EvMonitoring,
	EvExitCondition,
	EvBrokerCloseRequest,
	EvBrokerCloseConfirm,
	EvPositionClosed,
	EvFinalReconciliation,
}

type Event struct {
	Timestamp       time.Time `json:"timestamp"`
	Event           string    `json:"event"`
	StrategyTradeID string    `json:"strategy_trade_id"`
	SignalID        string    `json:"signal_id"`
	DealReference   string    `json:"deal_reference"`
	DealID          string    `json:"deal_id"`
	BrokerState     string    `json:"broker_state"`
	LocalState      string    `json:"local_state"`
	Severity        string    `json:"severity,omitempty"`
	Note            string    `json:"note,omitempty"`
}

type Latency struct {
	SignalToIntentMS      int64 `json:"signal_to_intent_ms"`
	IntentToRequestMS     int64 `json:"intent_to_request_ms"`
	RequestToConfirmMS    int64 `json:"request_to_confirm_ms"`
	ConfirmToResolvedMS   int64 `json:"confirm_to_resolved_ms"`
}

type Heartbeat struct {
	Timestamp     time.Time `json:"timestamp"`
	Bid           float64   `json:"bid"`
	Ask           float64   `json:"ask"`
	BrokerUPL     float64   `json:"broker_upl"`
	ExpectedUPL   float64   `json:"expected_upl"`
	ResidualUPL   float64   `json:"residual_upl"`
	SL            float64   `json:"sl"`
	TP            float64   `json:"tp"`
	PositionState string    `json:"position_state"`
}

type FinalRecord struct {
	StrategyTradeID     string  `json:"strategy_trade_id"`
	SignalID            string  `json:"signal_id"`
	DealReference       string  `json:"deal_reference"`
	DealID              string  `json:"deal_id"`
	Direction           string  `json:"direction"`
	EntryFill           float64 `json:"entry_fill"`
	EntryReference      float64 `json:"entry_reference"`
	EntrySlippageAbs    float64 `json:"entry_slippage_abs"`
	EntrySlippagePct    float64 `json:"entry_slippage_pct"`
	SpreadAtSignal      float64 `json:"spread_at_signal"`
	Size                float64 `json:"size"`
	StopLoss            float64 `json:"stop_loss"`
	TakeProfit          float64 `json:"take_profit"`
	ExpectedRiskUSD     float64 `json:"expected_risk_usd"`
	MoneyPerPriceUnit   float64 `json:"money_per_price_unit"`
	StopDistance        float64 `json:"stop_distance"`
	ExitReason          string  `json:"exit_reason"`
	GrossBrokerPnL      float64 `json:"gross_broker_pnl"`
	HoldingDuration     string  `json:"holding_duration"`
	TradeOutcome        string  `json:"trade_outcome"`
	OperationalOutcome  string  `json:"operational_outcome"`
	LocalPositions      int     `json:"local_positions"`
	BrokerPositions     int     `json:"broker_positions"`
	PendingConfirms     int     `json:"pending_confirms"`
	DuplicateOpens      int     `json:"duplicate_open_attempts"`
	ProtectiveOK        bool    `json:"protective_ok"`
	CountsAgree         bool    `json:"counts_agree"`
	LifecycleComplete   bool    `json:"lifecycle_complete"`
	Latency             Latency `json:"latency"`
}

type Recorder struct {
	mu           sync.Mutex
	root         string
	id           string
	signalID     string
	dealRef      string
	dealID       string
	events       []Event
	heartbeats   []Heartbeat
	pre          *PreSignalSnapshot
	intel        []IntelligenceContext
	tSignal      time.Time
	tIntent      time.Time
	tRequest     time.Time
	tConfirm     time.Time
	tResolved    time.Time
	duplicates   int
	halt         bool
	haltReason   string
	maxHeartbeats int
}

func NewRecorder(root, id string) *Recorder {
	return &Recorder{root: root, id: id, maxHeartbeats: 720}
}

func (r *Recorder) Dir() string {
	if r == nil {
		return ""
	}
	return filepath.Join(r.root, r.id)
}

func (r *Recorder) ID() string {
	if r == nil {
		return ""
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.id
}

func (r *Recorder) Bind(signalID, dealRef, dealID string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if signalID != "" {
		r.signalID = signalID
	}
	if dealRef != "" {
		r.dealRef = dealRef
	}
	if dealID != "" {
		r.dealID = dealID
	}
	if r.id == "" {
		r.id = TradeID(r.signalID, r.dealRef, r.dealID, "GOLD")
	}
}

func (r *Recorder) PersistPreSignal(s PreSignalSnapshot) error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.pre != nil && r.pre.Immutable {
		return fmt.Errorf("pre-signal snapshot is immutable")
	}
	s = FreezePreSignal(s)
	r.pre = &s
	r.id = s.StrategyTradeID
	r.signalID = s.SignalID
	r.tSignal = s.Timestamp
	return r.writeJSONLocked("pre_signal.json", s)
}

func (r *Recorder) AttachIntel(ctx IntelligenceContext) error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	ctx = NewObservational(ctx)
	r.intel = append(r.intel, ctx)
	return r.appendJSONLLocked("intelligence_context.jsonl", ctx)
}

func (r *Recorder) Append(ev Event) error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if ev.Timestamp.IsZero() {
		ev.Timestamp = time.Now().UTC()
	}
	if ev.StrategyTradeID == "" {
		ev.StrategyTradeID = r.id
	}
	if ev.SignalID == "" {
		ev.SignalID = r.signalID
	}
	if ev.DealReference == "" {
		ev.DealReference = r.dealRef
	}
	if ev.DealID == "" {
		ev.DealID = r.dealID
	}
	switch ev.Event {
	case EvOrderIntent:
		r.tIntent = ev.Timestamp
	case EvBrokerOpenRequest:
		r.tRequest = ev.Timestamp
	case EvBrokerConfirm:
		r.tConfirm = ev.Timestamp
	case EvPositionResolved:
		r.tResolved = ev.Timestamp
	}
	r.events = append(r.events, ev)
	return r.appendJSONLLocked("lifecycle.jsonl", ev)
}

func (r *Recorder) Sample(h Heartbeat) error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.maxHeartbeats <= 0 {
		r.maxHeartbeats = 720
	}
	if len(r.heartbeats) >= r.maxHeartbeats {
		r.heartbeats = r.heartbeats[1:]
	}
	if h.Timestamp.IsZero() {
		h.Timestamp = time.Now().UTC()
	}
	h.ResidualUPL = h.BrokerUPL - h.ExpectedUPL
	r.heartbeats = append(r.heartbeats, h)
	return r.appendJSONLLocked("position_samples.jsonl", h)
}

func (r *Recorder) NoteDuplicateOpen() {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.duplicates++
}

func (r *Recorder) DuplicateOpens() int {
	if r == nil {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.duplicates
}

func (r *Recorder) Halt(reason string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.halt = true
	r.haltReason = reason
}

func (r *Recorder) Halted() (bool, string) {
	if r == nil {
		return false, ""
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.halt, r.haltReason
}

func (r *Recorder) Events() []string {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, len(r.events))
	for i, e := range r.events {
		out[i] = e.Event
	}
	return out
}

func (r *Recorder) RawLatency() Latency {
	if r == nil {
		return Latency{}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return Latency{
		SignalToIntentMS:    millis(r.tSignal, r.tIntent),
		IntentToRequestMS:   millis(r.tIntent, r.tRequest),
		RequestToConfirmMS:  millis(r.tRequest, r.tConfirm),
		ConfirmToResolvedMS: millis(r.tConfirm, r.tResolved),
	}
}

func (r *Recorder) Finish(fin FinalRecord) error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	fin.StrategyTradeID = r.id
	fin.SignalID = r.signalID
	fin.DealReference = r.dealRef
	fin.DealID = r.dealID
	fin.DuplicateOpens = r.duplicates
	fin.Latency = Latency{
		SignalToIntentMS:    millis(r.tSignal, r.tIntent),
		IntentToRequestMS:   millis(r.tIntent, r.tRequest),
		RequestToConfirmMS:  millis(r.tRequest, r.tConfirm),
		ConfirmToResolvedMS: millis(r.tConfirm, r.tResolved),
	}
	names := make([]string, len(r.events))
	for i, e := range r.events {
		names[i] = e.Event
	}
	fin.LifecycleComplete = LifecycleComplete(names)
	return r.writeJSONLocked("final.json", fin)
}

func LifecycleOrdered(events []string) bool {
	want := 0
	for _, ev := range events {
		if want >= len(RequiredLifecycle) {
			return true
		}
		if ev == RequiredLifecycle[want] {
			want++
			continue
		}
		if ev == EvRiskRejected && want == 2 {
			return true
		}
	}
	return want == len(RequiredLifecycle)
}

func LifecycleComplete(events []string) bool {
	seen := map[string]bool{}
	for _, e := range events {
		seen[e] = true
	}
	if seen[EvRiskRejected] {
		return seen[EvSignalObserved] && seen[EvRiskEvaluated]
	}
	for _, req := range RequiredLifecycle {
		if req == EvRiskAccepted && seen[EvRiskRejected] {
			continue
		}
		if !seen[req] {
			return false
		}
	}
	return true
}

func OneSignalOneOrder(openAttempts int) bool {
	return openAttempts <= 1
}

func millis(from, to time.Time) int64 {
	if from.IsZero() || to.IsZero() || to.Before(from) {
		return 0
	}
	return to.Sub(from).Milliseconds()
}

func (r *Recorder) writeJSONLocked(name string, v any) error {
	if r.root == "" || r.id == "" {
		return nil
	}
	dir := filepath.Join(r.root, r.id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name), b, 0644)
}

func (r *Recorder) appendJSONLLocked(name string, v any) error {
	if r.root == "" || r.id == "" {
		return nil
	}
	dir := filepath.Join(r.root, r.id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(dir, name), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	return enc.Encode(v)
}

func ExitReasonFromEvidence(reason string) string {
	switch reason {
	case "TP", "SL", "STRATEGY_CLOSE", "SESSION_CLOSE", "MANUAL_SAFETY", "BROKER_ACTION":
		return reason
	case "UNKNOWN", "":
		return "UNKNOWN"
	default:
		return reason
	}
}

func FinalOperationallyTrusted(fin FinalRecord) bool {
	return fin.LifecycleComplete &&
		fin.CountsAgree &&
		fin.LocalPositions == 0 &&
		fin.BrokerPositions == 0 &&
		fin.PendingConfirms == 0 &&
		fin.DuplicateOpens == 0 &&
		fin.ProtectiveOK &&
		fin.OperationalOutcome != OpsFailed
}
