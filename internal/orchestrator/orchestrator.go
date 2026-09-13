package orchestrator

import (
	"sort"
	"strings"

	"aurumflow/internal/eligibility"
	"aurumflow/internal/opportunity"
	"aurumflow/internal/worlddomain"
	"aurumflow/internal/worldstate"
)

const Mode = "SHADOW"

type DecisionProposal struct {
	Market              string
	Direction           string
	AttentionScore      float64
	Coverage            float64
	EvidenceConfidence  float64
	Macro               string
	CapitalFlow         string
	Positioning         string
	CrossAsset          string
	Legacy              string
	Microstructure      string
	Confidence          float64
	ConfidenceKind      string
	WorldHash           string
	Blocking            []string
	PortfolioBlocks     []string
	Setup               worlddomain.SetupState
	Eligibility         worlddomain.ExecEligibility
	Decision            string
}

type Extra struct {
	HistoryKnown   bool
	HistoryPresent bool
	WorldHash      string
}

type DecisionOrchestrator struct {
	Mode string
}

func New() *DecisionOrchestrator { return &DecisionOrchestrator{Mode: Mode} }

func (d *DecisionOrchestrator) Propose(ws worldstate.WorldState, r opportunity.Ranked, legacy string, micro string) DecisionProposal {
	return d.ProposeFull(ws, r, legacy, micro, r.State.Eligibility, nil)
}

func (d *DecisionOrchestrator) ProposeFull(ws worldstate.WorldState, r opportunity.Ranked, legacy, micro string, elig worlddomain.ExecEligibility, portBlocks []string) DecisionProposal {
	return d.ProposeWith(ws, r, legacy, micro, elig, portBlocks, Extra{WorldHash: ws.Hash})
}

func (d *DecisionOrchestrator) ProposeWith(ws worldstate.WorldState, r opportunity.Ranked, legacy, micro string, elig worlddomain.ExecEligibility, portBlocks []string, extra Extra) DecisionProposal {
	if elig == "" {
		elig = r.State.Eligibility
	}
	elig = eligibility.Canonical(elig)
	hash := extra.WorldHash
	if hash == "" {
		hash = ws.Hash
	}
	p := DecisionProposal{
		Market: r.Market, AttentionScore: r.Score, Coverage: r.Coverage,
		EvidenceConfidence: r.EvidenceConfidence,
		Macro: r.State.MacroAlignment, CapitalFlow: r.State.CapitalFlowContext,
		Positioning: r.State.Positioning, CrossAsset: r.State.RelativeStrength,
		Legacy: legacy, Microstructure: micro,
		Confidence: 0, ConfidenceKind: "UNKNOWN",
		WorldHash: hash,
		Setup: worlddomain.SetupInsufficient, Eligibility: elig, Direction: "NONE",
		Decision: "WATCH", PortfolioBlocks: UniqueSorted(portBlocks),
	}
	if r.State.PriceTrend == "UP" || r.State.PriceTrend == "STRONG_UP" {
		p.Direction = "LONG_CANDIDATE"
	} else if r.State.PriceTrend == "DOWN" || r.State.PriceTrend == "STRONG_DOWN" {
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
		if extra.HistoryKnown && extra.HistoryPresent {
			p.Setup = worlddomain.SetupNoSetup
			p.Blocking = append(p.Blocking, "no_legacy_setup")
		} else {
			p.Setup = worlddomain.SetupInsufficient
			p.Blocking = append(p.Blocking, "insufficient_legacy_data")
		}
		p.Decision = "WATCH"
	} else {
		p.Setup = worlddomain.SetupPotential
		p.Decision = "SETUP"
	}
	if p.Setup == worlddomain.SetupPotential && (elig == worlddomain.EligDemo || elig == worlddomain.EligCalibrated) && len(p.PortfolioBlocks) == 0 {
		p.Decision = "DEMO_CANDIDATE"
	}
	if elig == worlddomain.EligBlocked || elig == worlddomain.EligLiveProhibited || len(p.PortfolioBlocks) > 0 {
		p.Decision = "BLOCKED"
	}
	if elig == worlddomain.EligAnalysis || elig == worlddomain.EligDiscovered || elig == worlddomain.EligSpecValid {
		if p.Setup == worlddomain.SetupPotential {
			p.Decision = "BLOCKED"
			p.Blocking = append(p.Blocking, "monetary spec not calibrated")
		}
	}
	p.Blocking = UniqueSorted(p.Blocking)
	p.PortfolioBlocks = UniqueSorted(p.PortfolioBlocks)
	return p
}

func (d *DecisionOrchestrator) CanMutateBroker() bool { return false }

func UniqueSorted(xs []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, x := range xs {
		x = strings.TrimSpace(x)
		if x == "" || seen[x] {
			continue
		}
		seen[x] = true
		out = append(out, x)
	}
	sort.Strings(out)
	return out
}
