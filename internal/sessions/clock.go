package sessions

import "time"

type Phase string

const (
	Asia       Phase = "ASIA"
	Europe     Phase = "EUROPE"
	US         Phase = "US"
	Overlap    Phase = "OVERLAP"
	Transition Phase = "TRANSITION"
)

type Schedule struct {
	AsiaOpenUTC    int
	AsiaCloseUTC   int
	EuropeOpenUTC  int
	EuropeCloseUTC int
	USOpenUTC      int
	USCloseUTC     int
}

func DefaultSchedule() Schedule {
	return Schedule{AsiaOpenUTC: 0, AsiaCloseUTC: 8, EuropeOpenUTC: 7, EuropeCloseUTC: 16, USOpenUTC: 13, USCloseUTC: 21}
}

func PhaseAt(t time.Time, s Schedule) Phase {
	h := t.UTC().Hour()
	asia := in(h, s.AsiaOpenUTC, s.AsiaCloseUTC)
	eu := in(h, s.EuropeOpenUTC, s.EuropeCloseUTC)
	us := in(h, s.USOpenUTC, s.USCloseUTC)
	n := 0
	if asia {
		n++
	}
	if eu {
		n++
	}
	if us {
		n++
	}
	if n >= 2 {
		return Overlap
	}
	if asia {
		return Asia
	}
	if eu {
		return Europe
	}
	if us {
		return US
	}
	return Transition
}

func in(h, a, b int) bool {
	if a == b {
		return false
	}
	if a < b {
		return h >= a && h < b
	}
	return h >= a || h < b
}

type HandoffSnapshot struct {
	At     time.Time
	Phase  Phase
	Event  string
	Note   string
}

func Handoff(prev, cur Phase, t time.Time) (HandoffSnapshot, bool) {
	if prev == cur {
		return HandoffSnapshot{}, false
	}
	ev := "session transition"
	switch {
	case prev == Asia && (cur == Europe || cur == Overlap):
		ev = "Asia close / Europe open"
	case (prev == Europe || prev == Overlap) && (cur == US || cur == Overlap) && cur != Asia:
		ev = "US open"
	case prev == US && cur != US && cur != Overlap:
		ev = "US close"
	}
	return HandoffSnapshot{At: t.UTC(), Phase: cur, Event: ev, Note: "descriptive session handoff, no causality"}, true
}
