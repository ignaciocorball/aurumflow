package featready

import (
	"aurumflow/internal/livesurface"
	"aurumflow/internal/mktwarmup"
	"aurumflow/internal/worlddomain"
)

const (
	PriceLive       = "PRICE_LIVE"
	MomentumReady   = "MOMENTUM_READY"
	VolatilityReady = "VOLATILITY_READY"
	CrossAssetReady = "CROSS_ASSET_READY"
	LegacyReady     = "LEGACY_READY"
	NotReady        = "NOT_READY"
)

type Snapshot struct {
	Market     string
	Price      string
	Momentum   string
	Volatility string
	CrossAsset string
	Legacy     string
	History    string
}

func Assess(market string, q livesurface.Quote, f livesurface.Features, h mktwarmup.History) Snapshot {
	s := Snapshot{Market: market, Price: NotReady, Momentum: NotReady, Volatility: NotReady, CrossAsset: NotReady, Legacy: NotReady, History: h.Status}
	if q.Live {
		s.Price = PriceLive
	}
	if q.Live && f.Ret1h != nil && f.Momentum != livesurface.MomUnknown && f.Momentum != livesurface.MomClosed && f.Momentum != livesurface.MomStale {
		s.Momentum = MomentumReady
	}
	if q.Live && f.Vol1h != nil && f.VolState != livesurface.VolUnknown {
		s.Volatility = VolatilityReady
	}
	if q.Live && f.RelStrength != "" && f.RelStrength != livesurface.MomUnknown {
		s.CrossAsset = CrossAssetReady
	}
	if h.Status == mktwarmup.StatusReady {
		s.Legacy = LegacyReady
	}
	return s
}

func Quality(q livesurface.Quote, f livesurface.Features) worlddomain.SensorHealth {
	if q.Stale {
		return worlddomain.HealthStale
	}
	if !q.Live {
		if q.MarketStatus == "" {
			return worlddomain.HealthUnavailable
		}
		return worlddomain.HealthDegraded
	}
	return worlddomain.HealthHealthy
}
