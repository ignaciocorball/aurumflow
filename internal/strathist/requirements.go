package strathist

import (
	"time"

	"aurumflow/config"
)

const (
	HistoryLoading = "HISTORY_LOADING"
	HistoryPartial = "HISTORY_PARTIAL"
	HistoryReady   = "HISTORY_READY"
	HistoryFailed  = "HISTORY_FAILED"
)

// LegacyRequirements are the actual loop/composer gates, not documentation guesses.
type LegacyRequirements struct {
	M5Required  int
	M15Required int
	H1Required  int
	H4Required  int
	M5Hard      bool
	M15Hard     bool
	H1Hard      bool
	H4Hard      bool
	M5Period    time.Duration
	M15Period   time.Duration
	H1Period    time.Duration
	H4Period    time.Duration
	M5Lookback  time.Duration
	M15Lookback time.Duration
	H1Lookback  time.Duration
	H4Lookback  time.Duration
}

func RequirementsFromConfig(cfg *config.Config) LegacyRequirements {
	r := LegacyRequirements{
		M15Required: 15, M15Hard: true, M15Period: 15 * time.Minute, M15Lookback: 7 * 24 * time.Hour,
		M5Period: 5 * time.Minute, M5Lookback: 48 * time.Hour,
		H1Period: time.Hour, H1Lookback: 14 * 24 * time.Hour,
		H4Period: 4 * time.Hour, H4Lookback: 21 * 24 * time.Hour,
	}
	if cfg != nil && cfg.M5Refiner != nil && cfg.M5Refiner.Enabled {
		r.M5Required = 10
		r.M5Hard = false
	}
	if cfg != nil && cfg.Strategy.UseH1Filter {
		r.H1Required = 10
		r.H1Hard = false
	}
	if cfg != nil && cfg.Strategy.UseH4Filter {
		r.H4Required = 60
		r.H4Hard = true
	}
	return r
}

func Status(m5, m15, h1, h4 int, req LegacyRequirements, fetchErr bool) string {
	if fetchErr && m15 == 0 && m5 == 0 && h1 == 0 && h4 == 0 {
		return HistoryFailed
	}
	if m15 == 0 && m5 == 0 && h1 == 0 && h4 == 0 {
		return HistoryLoading
	}
	if req.M15Hard && m15 < req.M15Required {
		return HistoryPartial
	}
	if req.H4Hard && h4 < req.H4Required {
		return HistoryPartial
	}
	if req.M5Hard && m5 < req.M5Required {
		return HistoryPartial
	}
	if req.H1Hard && h1 < req.H1Required {
		return HistoryPartial
	}
	if fetchErr {
		return HistoryPartial
	}
	return HistoryReady
}
