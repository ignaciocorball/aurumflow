package livesurface

import (
	"testing"
	"time"
)

func TestClosedQuoteDoesNotActLive(t *testing.T) {
	now := time.Date(2026, 9, 13, 18, 0, 0, 0, time.UTC)
	q := NewQuote("US100", "US100", 20000, 20002, "CLOSED", now.Add(-time.Minute), now, 2*time.Minute)
	if q.Live {
		t.Fatal("closed quote treated live")
	}
	hist := []Candle{
		{Time: now.Add(-2 * time.Hour), Close: 19800},
		{Time: now.Add(-time.Hour), Close: 19900},
		{Time: now, Close: 20000},
	}
	f := FeaturesFrom("US100", q, hist, now)
	if f.Momentum == MomUp || f.Momentum == MomStrongUp {
		t.Fatalf("closed momentum ranked live: %s", f.Momentum)
	}
	if f.Momentum != MomClosed {
		t.Fatalf("want CLOSED got %s", f.Momentum)
	}
}

func TestLiveQuoteFreshness(t *testing.T) {
	now := time.Date(2026, 9, 13, 18, 0, 0, 0, time.UTC)
	fresh := NewQuote("GOLD", "GOLD", 3600, 3601, "TRADEABLE", now.Add(-10*time.Second), now, 2*time.Minute)
	if !fresh.Live || fresh.Stale {
		t.Fatal(fresh)
	}
	stale := NewQuote("GOLD", "GOLD", 3600, 3601, "TRADEABLE", now.Add(-10*time.Minute), now, 2*time.Minute)
	if stale.Live || !stale.Stale {
		t.Fatal(stale)
	}
}

func TestMomentumAndVolatilityDeterministic(t *testing.T) {
	up := 0.008
	up4 := 0.02
	if MomentumOf(&up, &up4) != MomStrongUp {
		t.Fatal(MomentumOf(&up, &up4))
	}
	flat := 0.0001
	if MomentumOf(&flat, &flat) != MomFlat {
		t.Fatal(MomentumOf(&flat, &flat))
	}
	cur, hist := 0.01, 0.01
	if VolatilityOf(&cur, &hist) != VolNormal {
		t.Fatal(VolatilityOf(&cur, &hist))
	}
	if VolatilityOf(nil, &hist) != VolUnknown {
		t.Fatal("nil vol")
	}
}

func TestRelativeStrengthIgnoresClosed(t *testing.T) {
	g, s := 0.01, -0.01
	rs := RelativeStrength(map[string]*float64{"GOLD": &g, "SILVER": &s}, map[string]bool{"GOLD": true, "SILVER": false})
	if rs["SILVER"] != MomUnknown {
		t.Fatal(rs)
	}
	if rs["GOLD"] != MomFlat && rs["GOLD"] != MomUp {
		t.Fatal(rs["GOLD"])
	}
}

func TestBreadthUnknownWhenClosed(t *testing.T) {
	u1 := 0.01
	b := BreadthOf(&u1, &u1, &u1, map[string]*float64{"AAPL": &u1}, false)
	if b.State != BreadthUnknown {
		t.Fatal(b)
	}
	b = BreadthOf(&u1, &u1, &u1, map[string]*float64{"AAPL": &u1, "MSFT": &u1, "NVDA": &u1, "META": &u1}, true)
	if b.State != BreadthStrongBroad && b.State != BreadthBroad {
		t.Fatal(b)
	}
}

func TestRollingCorrInsufficient(t *testing.T) {
	if RollingCorr([]Candle{{Close: 1}}, []Candle{{Close: 1}}, 8) != nil {
		t.Fatal("emitted without coverage")
	}
}
