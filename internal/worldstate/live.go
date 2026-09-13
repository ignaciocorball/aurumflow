package worldstate

import (
	"strings"

	"aurumflow/internal/livesurface"
	"aurumflow/internal/sessions"
	"aurumflow/internal/worlddomain"
)

func ApplyLiveFrame(ws WorldState, frame livesurface.Frame) WorldState {
	ws.Live = frame
	if ws.Markets == nil {
		ws.Markets = map[string]MarketState{}
	}
	for id, q := range frame.Quotes {
		st := ws.Markets[id]
		if st.Market == "" {
			st.Market = id
			st.Resolved = true
			st.Eligibility = worlddomain.EligAnalysis
			st.PriceTrend = "UNKNOWN"
			st.RelativeStrength = "UNKNOWN"
			st.Volatility = "UNKNOWN"
			st.CapitalFlowContext = "UNKNOWN"
			st.Positioning = "UNKNOWN"
			st.MacroAlignment = "UNKNOWN"
			st.CrossAsset = "UNKNOWN"
			st.DataQuality = worlddomain.HealthUnknown
		}
		st.Bid, st.Ask, st.Mid = q.Bid, q.Ask, q.Mid
		st.MarketStatus = q.MarketStatus
		st.QuoteAgeSec = q.Age.Seconds()
		st.SessionLocal = sessions.MarketHours(id, q.MarketStatus, ws.AsOf)
		st.SourceBadge = "LIVE_MARKET"
		feat := frame.Features[id]
		if q.Stale {
			st.DataQuality = worlddomain.HealthDegraded
			st.PriceTrend = livesurface.MomStale
			st.RelativeStrength = livesurface.MomUnknown
			st.Volatility = livesurface.VolUnknown
		} else if !q.Live {
			st.DataQuality = worlddomain.HealthDegraded
			st.PriceTrend = livesurface.MomClosed
			st.RelativeStrength = livesurface.MomUnknown
			st.Volatility = livesurface.VolUnknown
		} else {
			st.DataQuality = worlddomain.HealthHealthy
			if feat.Momentum != "" {
				st.PriceTrend = feat.Momentum
			}
			if feat.RelStrength != "" {
				st.RelativeStrength = feat.RelStrength
			}
			if feat.VolState != "" {
				st.Volatility = feat.VolState
			}
			st.Evidence = appendUnique(st.Evidence, "live Capital quote observed")
		}
		if id == "OIL_CRUDE" || id == "OIL" {
			st.MacroAlignment = "UNKNOWN"
			st.Evidence = appendUnique(st.Evidence, "physical EIA remains UNKNOWN; price is not physical tightening")
		}
		ws.Markets[id] = st
	}
	if frame.Precious.Ratio != nil {
		if g, ok := ws.Markets["GOLD"]; ok {
			g.CrossAsset = frame.Precious.GoldRS
			g.Evidence = appendUnique(g.Evidence, "gold/silver live ratio observed")
			ws.Markets["GOLD"] = g
		}
		if s, ok := ws.Markets["SILVER"]; ok {
			s.CrossAsset = frame.Precious.SilverRS
			ws.Markets["SILVER"] = s
		}
	}
	if ac, ok := ws.AssetClasses[worlddomain.AssetEquities]; ok && frame.Breadth.State != "" && frame.Breadth.State != livesurface.BreadthUnknown {
		ac.Momentum = frame.Breadth.State
		ws.AssetClasses[worlddomain.AssetEquities] = ac
	}
	ws.LiveAt = frame.AsOf
	if ws.LiveAt.IsZero() {
		ws.LiveAt = ws.AsOf
	}
	return Finalize(ws)
}

func appendUnique(xs []string, v string) []string {
	for _, x := range xs {
		if strings.EqualFold(x, v) {
			return xs
		}
	}
	return append(xs, v)
}
