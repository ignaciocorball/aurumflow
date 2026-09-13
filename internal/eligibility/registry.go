package eligibility

import (
	"aurumflow/internal/money"
	"aurumflow/internal/worlddomain"
)

const MultiMarketDemoEnv = "MULTI_MARKET_DEMO_EXECUTION"

type Entry struct {
	Canonical   string
	Epic        string
	Status      worlddomain.ExecEligibility
	SpecStatus  string
	Reason      string
	Tradeable   bool
	DemoHost    bool
}

type Registry struct {
	entries map[string]Entry
}

func New() *Registry { return &Registry{entries: map[string]Entry{}} }

func (r *Registry) Set(e Entry) {
	if e.Status == "" {
		e.Status = worlddomain.EligAnalysis
	}
	r.entries[e.Canonical] = e
}

func (r *Registry) Get(canonical string) Entry {
	e, ok := r.entries[canonical]
	if !ok {
		return Entry{Canonical: canonical, Status: worlddomain.EligAnalysis, Reason: "not discovered"}
	}
	return e
}

func (r *Registry) All() []Entry {
	out := make([]Entry, 0, len(r.entries))
	for _, e := range r.entries {
		out = append(out, e)
	}
	return out
}

func Advance(e Entry, spec money.MonetaryInstrumentSpec, calibrated bool, live bool) Entry {
	if live {
		e.Status = worlddomain.EligLiveProhibited
		e.Reason = "LIVE impossible"
		return e
	}
	if e.Epic == "" {
		e.Status = worlddomain.EligAnalysis
		e.Reason = "ANALYSIS_ONLY"
		return e
	}
	e.Status = worlddomain.EligDiscovered
	e.Reason = "DEMO_DISCOVERED"
	if spec.Epic == "" {
		spec.Epic = e.Epic
	}
	e.SpecStatus = spec.ValidationStatus
	if spec.DealingComplete() && spec.ValidationStatus != money.Unverified {
		e.Status = worlddomain.EligSpecValid
		e.Reason = "DEMO_SPEC_VALID"
	}
	if calibrated && spec.ValidationStatus == money.RuntimeValidated {
		e.Status = worlddomain.EligCalibrated
		e.Reason = "DEMO_CALIBRATED"
	}
	if e.Status == worlddomain.EligCalibrated && e.Tradeable && e.DemoHost {
		e.Status = worlddomain.EligDemo
		e.Reason = "DEMO_ELIGIBLE"
	}
	if spec.MoneyPerPriceUnit <= 0 && calibrated {
		e.Status = worlddomain.EligBlocked
		e.Reason = "unknown monetary risk"
	}
	return e
}

func LiveAlwaysProhibited() worlddomain.ExecEligibility {
	return worlddomain.EligLiveProhibited
}
