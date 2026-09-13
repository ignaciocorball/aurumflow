package orchestrator

import (
	"aurumflow/internal/opportunity"
	"aurumflow/internal/worlddomain"
	"aurumflow/internal/worldstate"
)

const Mode = "SHADOW"

type DecisionProposal struct {
	Market          string
	Direction       string
	AttentionScore  float64
	Coverage        float64
	Macro           string
	CapitalFlow     string
	Positioning     string
	CrossAsset      string
	Legacy          string
	Microstructure  string
	Confidence      float64
	Blocking        []string
	PortfolioBlocks []string
	Setup           worlddomain.SetupState
	Eligibility     worlddomain.ExecEligibility
	Decision        string
}

type DecisionOrchestrator struct {
	Mode string
}

func New() *DecisionOrchestrator { return &DecisionOrchestrator{Mode: Mode} }

func (d *DecisionOrchestrator) Propose(ws worldstate.WorldState, r opportunity.Ranked, legacy string, micro string) DecisionProposal {
	return d.ProposeFull(ws, r, legacy, micro, r.State.Eligibility, nil)
}

func (d *DecisionOrchestrator) ProposeFull(ws worldstate.WorldState, r opportunity.Ranked, legacy, micro string, elig worlddomain.ExecEligibility, portBlocks []string) DecisionProposal {
	if elig == "" {
		elig = r.State.Eligibility
	}
	p := DecisionProposal{
		Market: r.Market, AttentionScore: r.Score, Coverage: r.Coverage,
		Macro: r.State.MacroAlignment, CapitalFlow: r.State.CapitalFlowContext,
		Positioning: r.State.Positioning, CrossAsset: r.State.RelativeStrength,
		Legacy: legacy, Microstructure: micro, Confidence: ws.Confidence,
		Setup: worlddomain.SetupNone, Eligibility: elig, Direction: "NONE",
		Decision: "WATCH", PortfolioBlocks: append([]string{}, portBlocks...),
	}
	if r.State.PriceTrend == "UP" {
		p.Direction = "LONG_CANDIDATE"
	} else if r.State.PriceTrend == "DOWN" {
		p.Direction = "SHORT_CANDIDATE"
	}
	if micro == "" && !r.State.MicroAvailable {
		p.Microstructure = "UNAVAILABLE"
	}
	if elig != worlddomain.EligDemo && elig != worlddomain.EligCalibrated {
		p.Blocking = append(p.Blocking, string(elig))
	}
	if r.State.DataQuality != worlddomain.HealthHealthy && r.State.DataQuality != "" {
		p.Blocking = append(p.Blocking, "data_quality")
	}
	if legacy == "" {
		p.Blocking = append(p.Blocking, "no_legacy_setup")
		p.Decision = "WATCH"
	} else {
		p.Setup = worlddomain.SetupPotential
		p.Decision = "SETUP"
	}
	if p.Setup == worlddomain.SetupPotential && (elig == worlddomain.EligDemo || elig == worlddomain.EligCalibrated) && len(portBlocks) == 0 {
		p.Decision = "DEMO_CANDIDATE"
	}
	if elig == worlddomain.EligBlocked || elig == worlddomain.EligLiveProhibited || len(portBlocks) > 0 {
		p.Decision = "BLOCKED"
		p.PortfolioBlocks = append(p.PortfolioBlocks, portBlocks...)
	}
	if elig == worlddomain.EligNotCalibrated || elig == worlddomain.EligAnalysis || elig == worlddomain.EligDiscovered || elig == worlddomain.EligSpecValid {
		if p.Setup == worlddomain.SetupPotential {
			p.Decision = "BLOCKED"
			p.Blocking = append(p.Blocking, "monetary spec not calibrated")
		}
	}
	return p
}

func (d *DecisionOrchestrator) CanMutateBroker() bool { return false }
