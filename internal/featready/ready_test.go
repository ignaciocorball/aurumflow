package featready

import (
	"testing"
	"time"

	"aurumflow/internal/livesurface"
	"aurumflow/internal/mktwarmup"
	"aurumflow/internal/worlddomain"
)

func TestNoFeatureOnInsufficientHistory(t *testing.T) {
	q := livesurface.NewQuote("GOLD", "GOLD", 1, 2, "CLOSED", time.Now().UTC(), time.Now().UTC(), time.Minute)
	f := livesurface.Features{Momentum: livesurface.MomClosed}
	s := Assess("GOLD", q, f, mktwarmup.History{Status: mktwarmup.StatusWarming})
	if s.Momentum == MomentumReady || s.Legacy == LegacyReady {
		t.Fatal(s)
	}
}

func TestLiveFeatures(t *testing.T) {
	now := time.Now().UTC()
	q := livesurface.NewQuote("BTC", "BTCUSD", 1, 2, "TRADEABLE", now, now, time.Minute)
	r := 0.01
	f := livesurface.Features{Ret1h: &r, Momentum: livesurface.MomUp, Vol1h: &r, VolState: livesurface.VolNormal, RelStrength: livesurface.MomUp, Live: true}
	s := Assess("BTC", q, f, mktwarmup.History{Status: mktwarmup.StatusReady, M5: 60, H1: 24, H4: 12})
	if s.Price != PriceLive || s.Momentum != MomentumReady || s.Legacy != LegacyReady {
		t.Fatal(s)
	}
}

func TestCapitalFailNotLiveRanking(t *testing.T) {
	q := livesurface.Quote{Market: "US100"}
	f := livesurface.Features{}
	s := Assess("US100", q, f, mktwarmup.History{})
	if s.Momentum == MomentumReady || s.Price == PriceLive {
		t.Fatal("invented live features after capital fail")
	}
	if Quality(q, f) != worlddomain.HealthUnavailable {
		t.Fatal(Quality(q, f))
	}
}
