package scanner

import (
	"time"

	"aurumflow/internal/instrument"
	"aurumflow/internal/opportunity"
	"aurumflow/internal/worlddomain"
	"aurumflow/internal/worldstate"
)

type Watch struct {
	Market string
	Tier   worlddomain.AttentionTier
	Mode   string
}

type Activation struct {
	Market string
	Tier   worlddomain.AttentionTier
	Legacy bool
	Micro  bool
	Quotes bool
	Slow   bool
}

type JournalEntry struct {
	At      time.Time
	Market  string
	Action  string
	Epic    string
}

type Scanner struct {
	Watched []Watch
	Journal []JournalEntry
}

func (s *Scanner) IngestDiscovery(maps []instrument.Mapping, at time.Time) {
	for _, m := range maps {
		if m.Epic == "" || m.Canonical == "" {
			continue
		}
		s.Watched = append(s.Watched, Watch{Market: m.Canonical, Tier: worlddomain.TierC, Mode: "ANALYSIS_ONLY"})
		s.Journal = append(s.Journal, JournalEntry{At: at.UTC(), Market: m.Canonical, Action: "discovered", Epic: m.Epic})
	}
}

func (s *Scanner) ApplyRanks(ranks []opportunity.Ranked) []Activation {
	var out []Activation
	s.Watched = s.Watched[:0]
	for _, r := range ranks {
		mode := "ANALYSIS_ONLY"
		if r.Tier == worlddomain.TierA {
			mode = "DEEP_WATCH"
		}
		s.Watched = append(s.Watched, Watch{Market: r.Market, Tier: r.Tier, Mode: mode})
		a := Activation{Market: r.Market, Tier: r.Tier, Slow: true}
		switch r.Tier {
		case worlddomain.TierA:
			a.Quotes, a.Legacy, a.Micro = true, true, r.State.MicroAvailable
		case worlddomain.TierB:
			a.Quotes = true
		}
		out = append(out, a)
	}
	return out
}

func (s *Scanner) CanTrade() bool { return false }

func AttachOpportunity(ws worldstate.WorldState, ranks []opportunity.Ranked) worldstate.WorldState {
	ws.Opportunity = nil
	for _, r := range ranks {
		st := r.State
		st.Attention = r.Score
		st.Coverage = r.Coverage
		ws.Opportunity = append(ws.Opportunity, st)
	}
	return ws
}
