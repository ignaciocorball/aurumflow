package stratrade

import (
	"strings"
	"time"

	"aurumflow/internal/strategy"
)

type SessionReport struct {
	BrokerMarket     string
	StrategySession  string
	StrategyReady    bool
	StrategyWaiting  string
	NextSession      string
	NextSessionAt    time.Time
	Allowed          []string
}

func ObserveSession(now time.Time, allowed []string, brokerStatus string) SessionReport {
	r := SessionReport{
		BrokerMarket:    strings.ToUpper(strings.TrimSpace(brokerStatus)),
		StrategySession: strategy.GetSessionInfo(now),
		Allowed:         append([]string{}, allowed...),
	}
	if r.BrokerMarket == "" {
		r.BrokerMarket = "UNKNOWN"
	}
	can := strategy.CanTrade(now, allowed)
	r.StrategyReady = can && (r.BrokerMarket == "TRADEABLE" || r.BrokerMarket == "ON")
	if !can {
		r.StrategyWaiting = "STRATEGY WAITING FOR SESSION"
	} else if r.BrokerMarket != "TRADEABLE" && r.BrokerMarket != "ON" {
		r.StrategyWaiting = "STRATEGY WAITING FOR BROKER"
	} else {
		r.StrategyWaiting = "STRATEGY WAITING"
	}
	nextAllowed := allowed
	if len(nextAllowed) == 0 {
		nextAllowed = []string{strategy.SessionAll}
	}
	if name, at, ok := strategy.NextSessionStart(now, nextAllowed); ok {
		r.NextSession = name
		r.NextSessionAt = at
	}
	return r
}
