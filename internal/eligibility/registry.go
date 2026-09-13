package eligibility

import (
	"sort"
	"strings"

	"aurumflow/internal/money"
	"aurumflow/internal/worlddomain"
	"aurumflow/internal/worldstate"
)

const MultiMarketDemoEnv = "MULTI_MARKET_DEMO_EXECUTION"

// Canonical lifecycle. LIVE is never a tradable state; it maps to BLOCKED.
//
//	ANALYSIS_ONLY → DEMO_DISCOVERED → DEMO_SPEC_VALID → DEMO_CALIBRATED → DEMO_ELIGIBLE
//	any state → BLOCKED
//
// No jumps. prepare advances DISCOVERED→SPEC_VALID. calibrate advances SPEC_VALID→CALIBRATED.
// risk/safety checks advance CALIBRATED→ELIGIBLE. DEMO_NOT_CALIBRATED is not emitted.
const (
	DocLifecycle = `
ANALYSIS_ONLY
  discover unique Capital identity
DEMO_DISCOVERED
  prepare (read-only monetary spec)
DEMO_SPEC_VALID
  explicit calibrate command
DEMO_CALIBRATED
  tradeable + demo host + known monetary risk
DEMO_ELIGIBLE
any → BLOCKED (LIVE, unknown money, failed gate)
`
)

type Entry struct {
	Canonical  string
	Epic       string
	Status     worlddomain.ExecEligibility
	SpecStatus string
	Reason     string
	Tradeable  bool
	DemoHost   bool
}

type Registry struct {
	entries map[string]Entry
}

func New() *Registry { return &Registry{entries: map[string]Entry{}} }

func (r *Registry) Set(e Entry) {
	if e.Status == "" {
		e.Status = worlddomain.EligAnalysis
	}
	if e.Status == worlddomain.EligNotCalibrated {
		e.Status = worlddomain.EligDiscovered
		if e.Reason == "" {
			e.Reason = "DEMO_DISCOVERED"
		}
	}
	r.entries[e.Canonical] = e
}

func (r *Registry) Get(canonical string) Entry {
	e, ok := r.entries[canonical]
	if !ok {
		return Entry{Canonical: canonical, Status: worlddomain.EligAnalysis, Reason: "ANALYSIS_ONLY"}
	}
	return e
}

func (r *Registry) All() []Entry {
	keys := make([]string, 0, len(r.entries))
	for k := range r.entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]Entry, 0, len(keys))
	for _, k := range keys {
		out = append(out, r.entries[k])
	}
	return out
}

func rank(s worlddomain.ExecEligibility) int {
	switch s {
	case worlddomain.EligAnalysis:
		return 0
	case worlddomain.EligDiscovered:
		return 1
	case worlddomain.EligSpecValid:
		return 2
	case worlddomain.EligCalibrated:
		return 3
	case worlddomain.EligDemo:
		return 4
	case worlddomain.EligBlocked, worlddomain.EligLiveProhibited:
		return 90
	default:
		return 0
	}
}

func Allowed(from, to worlddomain.ExecEligibility) bool {
	if to == worlddomain.EligBlocked || to == worlddomain.EligLiveProhibited {
		return true
	}
	if from == to {
		return true
	}
	if from == worlddomain.EligBlocked || from == worlddomain.EligLiveProhibited {
		return false
	}
	return rank(to) == rank(from)+1
}

func ApplyTransition(e Entry, to worlddomain.ExecEligibility, reason string) Entry {
	if !Allowed(e.Status, to) {
		e.Reason = "illegal transition " + string(e.Status) + " → " + string(to)
		return e
	}
	e.Status = to
	e.Reason = reason
	return e
}

func Advance(e Entry, spec money.MonetaryInstrumentSpec, calibrated bool, live bool) Entry {
	if live {
		e.Status = worlddomain.EligBlocked
		e.Reason = "LIVE_PROHIBITED"
		return e
	}
	if e.Status == "" {
		e.Status = worlddomain.EligAnalysis
	}
	if e.Epic == "" {
		e.Status = worlddomain.EligAnalysis
		e.Reason = "ANALYSIS_ONLY"
		return e
	}
	if spec.Epic == "" {
		spec.Epic = e.Epic
	}
	e.SpecStatus = spec.ValidationStatus
	e = step(e, worlddomain.EligDiscovered, "DEMO_DISCOVERED")
	if spec.DealingComplete() && spec.ValidationStatus != money.Unverified {
		e = step(e, worlddomain.EligSpecValid, "DEMO_SPEC_VALID")
	}
	if calibrated && spec.ValidationStatus == money.RuntimeValidated {
		e = step(e, worlddomain.EligCalibrated, "DEMO_CALIBRATED")
	}
	if e.Status == worlddomain.EligCalibrated && e.Tradeable && e.DemoHost {
		e = step(e, worlddomain.EligDemo, "DEMO_ELIGIBLE")
	}
	if spec.MoneyPerPriceUnit <= 0 && calibrated {
		e.Status = worlddomain.EligBlocked
		e.Reason = "unknown monetary risk"
	}
	return e
}

func step(e Entry, to worlddomain.ExecEligibility, reason string) Entry {
	if e.Status == to {
		e.Reason = reason
		return e
	}
	if Allowed(e.Status, to) {
		e.Status = to
		e.Reason = reason
	}
	return e
}

func LiveAlwaysProhibited() worlddomain.ExecEligibility {
	return worlddomain.EligBlocked
}

func (r *Registry) ApplyToWorld(ws worldstate.WorldState) worldstate.WorldState {
	if ws.Markets == nil {
		return worldstate.Finalize(ws)
	}
	for id, st := range ws.Markets {
		e := r.Get(id)
		st.Eligibility = e.Status
		st.EligReason = e.Reason
		ws.Markets[id] = st
	}
	return worldstate.Finalize(ws)
}

func Canonical(s worlddomain.ExecEligibility) worlddomain.ExecEligibility {
	switch s {
	case worlddomain.EligNotCalibrated:
		return worlddomain.EligDiscovered
	case worlddomain.EligLiveProhibited:
		return worlddomain.EligBlocked
	case "":
		return worlddomain.EligAnalysis
	default:
		return s
	}
}

func IsCanonical(s worlddomain.ExecEligibility) bool {
	switch Canonical(s) {
	case worlddomain.EligAnalysis, worlddomain.EligDiscovered, worlddomain.EligSpecValid,
		worlddomain.EligCalibrated, worlddomain.EligDemo, worlddomain.EligBlocked:
		return true
	default:
		return false
	}
}

func SameAuthority(market worlddomain.ExecEligibility, proposal worlddomain.ExecEligibility) bool {
	return Canonical(market) == Canonical(proposal)
}

func SortMarkets(ids []string) []string {
	out := append([]string{}, ids...)
	sort.Slice(out, func(i, j int) bool { return strings.ToUpper(out[i]) < strings.ToUpper(out[j]) })
	return out
}
