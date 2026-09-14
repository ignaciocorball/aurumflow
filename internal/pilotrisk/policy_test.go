package pilotrisk

import (
	"strings"
	"testing"

	"aurumflow/internal/portfoliorisk"
)

func base(equity float64) Input {
	return Input{
		Market: "US100", Equity: equity, StopDistance: 40, MPU: 1,
		MinDealSize: 0.001, SizeStep: 0.001, MaxDealSize: 1250,
	}
}

func TestPilotCapsFromActualEquity(t *testing.T) {
	if PerPositionCap(1000) != 5 || AggregateCap(1000) != 10 {
		t.Fatal(PerPositionCap(1000), AggregateCap(1000))
	}
	if EffectiveAggregate(100000, 300) != 300 {
		t.Fatal("hard aggregate must bind")
	}
	if EffectivePerPosition(100000, 150) != 150 {
		t.Fatal("hard class must bind")
	}
}

func TestMinSizeExceedsPilotRisk(t *testing.T) {
	in := base(1000)
	in.MinDealSize = 1
	in.StopDistance = 20
	d := Evaluate(in)
	if d.Pass || d.Block != ReasonMinSize {
		t.Fatalf("%+v", d)
	}
}

func TestAggregateOnePercent(t *testing.T) {
	in := base(1000)
	in.OpenRisk = 6
	d := Evaluate(in)
	if !d.Pass {
		t.Fatal(d.Block)
	}
	if in.OpenRisk+d.PlannedRisk > 10+1e-6 {
		t.Fatalf("aggregate %v", d)
	}
	in.OpenRisk = 10
	if Evaluate(in).Pass {
		t.Fatal("full aggregate")
	}
}

func TestGoldAndUS100Allowed(t *testing.T) {
	in := base(2000)
	in.OpenStrategy = 1
	in.OpenRisk = 8
	in.GroupOpen = map[string]int{portfoliorisk.GroupPrecious: 1}
	if !Evaluate(in).Pass {
		t.Fatal(Evaluate(in).Block)
	}
}

func TestUS100PlusUS500Prohibited(t *testing.T) {
	in := base(2000)
	in.Market = "US500"
	in.GroupOpen = map[string]int{portfoliorisk.GroupUSEquity: 1}
	d := Evaluate(in)
	if d.Pass || !strings.Contains(d.Block, portfoliorisk.GroupUSEquity) {
		t.Fatalf("%+v", d)
	}
	in.Market = "US30"
	if Evaluate(in).Pass {
		t.Fatal("US30")
	}
}

func TestMaxTwoGlobalAndTradeLimit(t *testing.T) {
	in := base(2000)
	in.OpenStrategy = 2
	if Evaluate(in).Block != ReasonMaxOpen {
		t.Fatal(Evaluate(in).Block)
	}
	in.OpenStrategy = 0
	in.TradesThisMarket = 1
	if Evaluate(in).Block != ReasonTradeLimit {
		t.Fatal(Evaluate(in).Block)
	}
}

func TestDoesNotChangeStopDistance(t *testing.T) {
	in := base(5000)
	stop := in.StopDistance
	_ = Evaluate(in)
	if in.StopDistance != stop {
		t.Fatal("overlay must not mutate stop")
	}
}
