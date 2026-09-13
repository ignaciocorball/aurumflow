package sessobs

import (
	"aurumflow/internal/sessions"
	"time"
)

// Observation is descriptive only. No strategy inference.
type Observation struct {
	At                 time.Time
	SourceSession      string
	DestinationSession string
	LeadershipBefore   string
	GoldState          string
	OilState           string
	RegionalEquity     string
	AfterHandoff       string
	Note               string
}

func Observe(prev, cur sessions.Phase, t time.Time, lead, gold, oil, regional string) (Observation, bool) {
	h, ok := sessions.Handoff(prev, cur, t)
	if !ok {
		return Observation{}, false
	}
	return Observation{
		At: h.At, SourceSession: string(prev), DestinationSession: string(cur),
		LeadershipBefore: lead, GoldState: gold, OilState: oil, RegionalEquity: regional,
		AfterHandoff: "UNKNOWN",
		Note:         "descriptive session handoff — continuation/reversal not inferred in P8.2",
	}, true
}
