package sessions

import (
	"strings"
	"time"
)

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

type StatusTransition struct {
	Market string
	From   string
	To     string
	At     time.Time
}

func BrokerTransition(market, prev, cur string, t time.Time) (StatusTransition, bool) {
	prev, cur = strings.ToUpper(strings.TrimSpace(prev)), strings.ToUpper(strings.TrimSpace(cur))
	if prev == cur || cur == "" {
		return StatusTransition{}, false
	}
	return StatusTransition{Market: market, From: prev, To: cur, At: t.UTC()}, true
}

const (
	MarketPreopen = "PREOPEN"
	MarketOpen    = "OPEN"
	MarketClosed  = "CLOSED"
	MarketUnknown = "UNKNOWN"
)

// MarketHours is descriptive clock state. Broker marketStatus remains authoritative for execution.
func MarketHours(canonical string, brokerStatus string, t time.Time) string {
	switch strings.ToUpper(strings.TrimSpace(brokerStatus)) {
	case "TRADEABLE", "OPEN", "ON":
		return MarketOpen
	case "CLOSED", "OFFLINE":
		return MarketClosed
	}
	if strings.EqualFold(canonical, "BTC") {
		return MarketOpen
	}
	h := t.UTC().Hour()
	switch strings.ToUpper(canonical) {
	case "J225", "CN50", "CHINA_HK", "HK":
		if h >= 0 && h < 8 {
			return MarketOpen
		}
		if h >= 23 {
			return MarketPreopen
		}
		return MarketClosed
	case "DE40", "UK100", "EUROPE", "UK":
		if h >= 7 && h < 16 {
			return MarketOpen
		}
		if h >= 6 && h < 7 {
			return MarketPreopen
		}
		return MarketClosed
	case "US100", "US500", "US30", "AAPL", "MSFT", "NVDA", "META", "AMZN":
		if h >= 13 && h < 21 {
			return MarketOpen
		}
		if h >= 12 && h < 13 {
			return MarketPreopen
		}
		return MarketClosed
	case "GOLD", "SILVER", "OIL", "OIL_CRUDE":
		if h >= 22 || h < 21 {
			return MarketOpen
		}
		return MarketClosed
	default:
		return MarketUnknown
	}
}
