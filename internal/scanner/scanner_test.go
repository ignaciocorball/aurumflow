package scanner

import (
	"testing"
	"time"

	"aurumflow/internal/instrument"
	"aurumflow/internal/opportunity"
	"aurumflow/internal/worlddomain"
	"aurumflow/internal/worldstate"
)

func TestScannerCannotTrade(t *testing.T) {
	s := &Scanner{}
	s.IngestDiscovery([]instrument.Mapping{{Canonical: "GOLD", Epic: "GOLD"}}, time.Now().UTC())
	if s.CanTrade() {
		t.Fatal("trade")
	}
	act := s.ApplyRanks([]opportunity.Ranked{{Market: "GOLD", Tier: worlddomain.TierA, State: worldstate.MarketState{Market: "GOLD"}}})
	if len(act) != 1 || act[0].Legacy != true {
		t.Fatal(act)
	}
}
