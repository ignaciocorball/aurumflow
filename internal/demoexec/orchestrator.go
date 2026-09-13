package demoexec

import (
	"os"
	"strings"

	"aurumflow/internal/orchestrator"
	"aurumflow/internal/worlddomain"
)

const EnvFlag = "MULTI_MARKET_DEMO_EXECUTION"

type Result struct {
	Accepted bool
	Reason   string
}

type Orchestrator struct {
	Enabled bool
}

func New() *Orchestrator {
	on := strings.EqualFold(os.Getenv(EnvFlag), "ON")
	return &Orchestrator{Enabled: on}
}

func (o *Orchestrator) Accept(p orchestrator.DecisionProposal, live bool) Result {
	if live {
		return Result{Reason: "LIVE fail-closed"}
	}
	if !o.Enabled {
		return Result{Reason: "MULTI_MARKET_DEMO_EXECUTION=OFF"}
	}
	if p.Decision != "DEMO_CANDIDATE" {
		return Result{Reason: "proposal is not DEMO_CANDIDATE"}
	}
	if p.Eligibility != worlddomain.EligDemo && p.Eligibility != worlddomain.EligCalibrated {
		return Result{Reason: "execution eligibility incomplete"}
	}
	if len(p.PortfolioBlocks) > 0 {
		return Result{Reason: "portfolio risk"}
	}
	return Result{Accepted: true, Reason: "gates passed; still not an order"}
}

func (o *Orchestrator) CanMutateBroker() bool { return false }
