package dataint

import "time"

const (
	Valid    = "VALID"
	Degraded = "DEGRADED"
	Invalid  = "INVALID"
)

type Interval struct {
	Start      time.Time `json:"start"`
	End        time.Time `json:"end"`
	Provider   string    `json:"provider"`
	EventTypes string    `json:"event_types"`
	Status     string    `json:"status"`
	Reason     string    `json:"reason"`
	Drops      int64     `json:"drops"`
	Gaps       int       `json:"gaps"`
	Resyncs    int       `json:"resyncs"`
}

func Classify(drops int64, gaps, resyncs int, bookSynced bool) Interval {
	in := Interval{Status: Valid, Reason: "ok", Drops: drops, Gaps: gaps, Resyncs: resyncs}
	if drops > 0 {
		in.Status = Degraded
		in.Reason = "BUS_BACKPRESSURE"
	}
	if gaps > 0 {
		in.Status = Degraded
		if in.Reason == "ok" {
			in.Reason = "BOOK_SEQUENCE_GAP"
		}
	}
	if !bookSynced && (drops > 0 || gaps > 0) {
		in.Status = Invalid
		in.Reason = "BOOK_UNSYNCED_WITH_LOSS"
	}
	return in
}
