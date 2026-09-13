package ops

import "time"

const (
	MaxSeriesPoints = 3600
	MaxTimeline     = 200
	SeriesHorizon   = time.Hour
)

type SeriesPoint struct {
	T        int64   `json:"t"`
	Price    float64 `json:"price"`
	Micro    float64 `json:"microprice"`
	Pressure float64 `json:"pressure"`
	DirP     float64 `json:"directional_pressure"`
	CVD      float64 `json:"cvd"`
	FlowEff  float64 `json:"flow_efficiency"`
	Impact   float64 `json:"impact_failure"`
	Synced   bool    `json:"book_synced"`
}

type TimelineEvent struct {
	T    int64  `json:"t"`
	Kind string `json:"kind"`
	Text string `json:"text"`
}

type BookLevel struct {
	Price float64 `json:"price"`
	Qty   float64 `json:"qty"`
}

func (s *Server) appendSeriesLocked(st Status, now time.Time) {
	if s.lastSample.IsZero() || now.Sub(s.lastSample) >= time.Second {
		s.lastSample = now
		s.series = append(s.series, SeriesPoint{
			T: now.UnixMilli(), Price: st.BTCPrice, Micro: st.Microprice,
			Pressure: st.Pressure, DirP: st.DirectionalP, CVD: st.CVD,
			FlowEff: st.FlowEfficiency, Impact: st.ImpactFailure, Synced: st.BookSynced,
		})
		cut := now.Add(-SeriesHorizon).UnixMilli()
		i := 0
		for i < len(s.series) && s.series[i].T < cut {
			i++
		}
		if i > 0 {
			s.series = s.series[i:]
		}
		if len(s.series) > MaxSeriesPoints {
			s.series = s.series[len(s.series)-MaxSeriesPoints:]
		}
	}
	s.noteLocked(st, now)
}

func (s *Server) noteLocked(st Status, now time.Time) {
	push := func(kind, text string) {
		if text == "" {
			return
		}
		s.timeline = append(s.timeline, TimelineEvent{T: now.UnixMilli(), Kind: kind, Text: text})
		if len(s.timeline) > MaxTimeline {
			s.timeline = s.timeline[len(s.timeline)-MaxTimeline:]
		}
	}
	if st.LastV1Class != "" && st.LastV1Class != s.prev.LastV1Class {
		push("v1", st.LastV1Class)
	}
	if st.AbsorptionStatus != "" && st.AbsorptionStatus != s.prev.AbsorptionStatus {
		push("absorption", st.AbsorptionStatus)
	}
	if st.LastLegacyDir != s.prev.LastLegacyDir && st.LastLegacyDir != 0 {
		if st.LastLegacyDir > 0 {
			push("legacy", "LEGACY LONG")
		} else {
			push("legacy", "LEGACY SHORT")
		}
	}
	if s.prev.BookSynced && !st.BookSynced {
		push("book", "book unsynced")
	}
	if st.BookGaps > s.prev.BookGaps {
		push("book", "book gap")
	}
	if st.Resyncs > s.prev.Resyncs {
		push("book", "book resync")
	}
	if st.KillSwitch && !s.prev.KillSwitch {
		push("safety", "kill switch ON")
	}
	if st.Reconnects > s.prev.Reconnects {
		push("feed", "provider reconnect")
	}
	if st.PositionsKnown && st.OpenPositions > s.prev.OpenPositions {
		push("position", "Position opened")
	}
	if st.PositionsKnown && s.prev.PositionsKnown && st.OpenPositions < s.prev.OpenPositions {
		push("position", "Position closed")
	}
	if st.LastExecution != "" && st.LastExecution != s.prev.LastExecution {
		push("capital", "Capital order")
	}
	prevM, curM := GoldMarketLabel(s.prev.MarketStatus), GoldMarketLabel(st.MarketStatus)
	if prevM != curM && curM == "TRADEABLE" {
		push("gold", "GOLD MARKET TRADEABLE")
	}
	s.prev = st
}

func (s *Server) Series() []SeriesPoint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]SeriesPoint, len(s.series))
	copy(out, s.series)
	return out
}

func (s *Server) Timeline() []TimelineEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]TimelineEvent, len(s.timeline))
	copy(out, s.timeline)
	return out
}
