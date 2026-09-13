package quotehealth

import (
	"strings"
	"time"

	"aurumflow/internal/livesurface"
	"aurumflow/internal/worlddomain"
)

type Snapshot struct {
	Market        string
	UpdatesPerMin float64
	Spread        float64
	SpreadPct     float64
	StaleCount    int
	AgeSec        float64
	Status        worlddomain.SensorHealth
}

func Of(q livesurface.Quote, updates int, window time.Duration, staleN int) Snapshot {
	s := Snapshot{Market: q.Market, Spread: q.Spread, StaleCount: staleN, AgeSec: q.Age.Seconds()}
	if q.Mid > 0 {
		s.SpreadPct = q.Spread / q.Mid
	}
	if window > 0 {
		s.UpdatesPerMin = float64(updates) / window.Minutes()
	}
	switch {
	case q.Stale:
		s.Status = worlddomain.HealthStale
	case !q.Live && strings.EqualFold(q.MarketStatus, "CLOSED"):
		s.Status = worlddomain.HealthDegraded
	case q.MarketStatus == "" && q.Mid <= 0:
		s.Status = worlddomain.HealthUnavailable
	case q.Live:
		s.Status = worlddomain.HealthHealthy
	default:
		s.Status = worlddomain.HealthDegraded
	}
	return s
}

func TapeState(q livesurface.Quote, histStatus string) string {
	if q.Stale {
		return "STALE"
	}
	if q.Live {
		return "LIVE"
	}
	if histStatus == "WARMING" {
		return "WARMING"
	}
	if strings.EqualFold(q.MarketStatus, "CLOSED") || q.MarketStatus != "" {
		return "CLOSED"
	}
	return "WARMING"
}
