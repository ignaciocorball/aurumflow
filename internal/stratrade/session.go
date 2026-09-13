package stratrade

import (
	"strings"
	"time"

	"aurumflow/internal/strategy"
)

type SessionReport struct {
	BrokerMarket     string
	StrategySession  string
	ConfigPolicy     string
	Eligible         bool
	Reason           string
	StrategyReady    bool
	StrategyWaiting  string
	NextSession      string
	NextSessionAt    time.Time
	Allowed          []string
}

func ObserveSession(now time.Time, allowed []string, brokerStatus string) SessionReport {
	ev := strategy.EvaluateSession(now, allowed)
	r := SessionReport{
		BrokerMarket:    strings.ToUpper(strings.TrimSpace(brokerStatus)),
		StrategySession: ev.ClockSession,
		ConfigPolicy:    ev.ConfigPolicy,
		Eligible:        ev.Eligible,
		Reason:          ev.Reason,
		Allowed:         append([]string{}, allowed...),
	}
	if r.BrokerMarket == "" {
		r.BrokerMarket = "UNKNOWN"
	}
	r.StrategyReady = ev.Eligible && (r.BrokerMarket == "TRADEABLE" || r.BrokerMarket == "ON")
	if !ev.Eligible {
		r.StrategyWaiting = "STRATEGY WAITING FOR SESSION"
	} else if r.BrokerMarket != "TRADEABLE" && r.BrokerMarket != "ON" {
		r.StrategyWaiting = "STRATEGY WAITING FOR BROKER"
	} else {
		r.StrategyWaiting = "STRATEGY WAITING"
	}
	r.NextSession, r.NextSessionAt = strategy.NextEligibleLabel(ev, now, allowed)
	return r
}
