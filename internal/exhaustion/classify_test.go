package exhaustion

import (
	"testing"
	"time"
)

func parseT(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

func timeMinute() time.Duration { return time.Minute }

func TestClassifyV1Frozen(t *testing.T) {
	if ClassifyV1(1, -15) != ClassExhaustion {
		t.Fatal("LONG + negative pressure")
	}
	if ClassifyV1(-1, 15) != ClassExhaustion {
		t.Fatal("SHORT + positive pressure")
	}
	if ClassifyV1(1, 15) != ClassContinuation {
		t.Fatal("LONG continuation")
	}
	if ClassifyV1(-1, -15) != ClassContinuation {
		t.Fatal("SHORT continuation")
	}
	if ClassifyV1(1, 14.9) != ClassNeutral {
		t.Fatal("below threshold")
	}
	if DirectionalPressure(1, -20) != -20 || DirectionalPressure(-1, 20) != -20 {
		t.Fatal("DP")
	}
	if ClassifyV1(1, 40) == ClassExhaustion {
		t.Fatal("must not invert PressureScore")
	}
}

func TestPastOnlyNoLookahead(t *testing.T) {
	t0 := parseT("2026-06-01T12:00:00Z")
	trades := []Trade{
		{T: t0.Add(-timeMinute()), Price: 100, Qty: 2, BuyerMaker: false},
		{T: t0, Price: 110, Qty: 99, BuyerMaker: false},
		{T: t0.Add(timeMinute()), Price: 120, Qty: 99, BuyerMaker: false},
	}
	past := PastOnly(trades, t0)
	if len(past) != 1 {
		t.Fatalf("future leaked %d", len(past))
	}
	fw := FlowFrom(trades, t0, timeMinute(), 0)
	if fw.BuyVol != 2 {
		t.Fatalf("must ignore t0 and later: %+v", fw)
	}
}

func TestEfficiencyPastOnly(t *testing.T) {
	e := NewEngine("BTCUSDT")
	t0 := parseT("2026-06-01T12:00:00Z")
	for i := 70; i >= 1; i-- {
		tt := t0.Add(-time.Duration(i) * time.Minute)
		e.OnTrade(tt, 100, 1, false)
		e.Observe(tt, 1, 6, 5, 1)
	}
	e.OnTrade(t0.Add(-30*time.Second), 100.01, 50, true)
	s := e.Observe(t0, 1, 6, -20, 1)
	if s.Classification != ClassExhaustion {
		t.Fatal(s.Classification)
	}
	if s.PressureScore != -20 || s.DirectionalPressure != -20 {
		t.Fatal("pressure mutated")
	}
	if s.Mode != ModeShadow || s.EvidenceStatus != "EXPERIMENTAL_RESEARCH_ONLY" {
		t.Fatal(s.Mode)
	}
	if s.Classification == AbsorptionNever || TradeOnlyAbsorption(s).Book.Available {
		t.Fatal("trade-only must not emit absorption")
	}
}

func TestReplayLiveIdenticalFeatures(t *testing.T) {
	t0 := parseT("2026-06-01T12:00:00Z")
	mk := func() Snapshot {
		e := NewEngine("BTCUSDT")
		e.OnTrade(t0.Add(-2*time.Minute), 100, 3, false)
		e.OnTrade(t0.Add(-time.Minute), 100.2, 4, true)
		return e.Observe(t0, 1, 6, -20, 1)
	}
	a, b := mk(), mk()
	if a.Features.FlowMagNorm != b.Features.FlowMagNorm || a.Features.ImpactFailure != b.Features.ImpactFailure {
		t.Fatalf("replay/live mismatch %+v %+v", a.Features, b.Features)
	}
	if a.Classification != ClassExhaustion {
		t.Fatal(a.Classification)
	}
}

func TestNoAbsorptionConstantFromEngine(t *testing.T) {
	e := NewEngine("BTCUSDT")
	s := e.Observe(time.Now().UTC(), 1, 6, -20, 1)
	if s.Classification == AbsorptionNever {
		t.Fatal("ABSORPTION_CONFIRMED forbidden")
	}
	if MayMutateBroker() {
		t.Fatal("exhaustion cannot mutate broker")
	}
}
