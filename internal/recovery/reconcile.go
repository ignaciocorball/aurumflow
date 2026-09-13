package recovery

import "strings"

const (
	Managed  = "MANAGED"
	Recovered = "RECOVERED"
	Unknown  = "UNKNOWN"
)

type BrokerPos struct {
	DealID, Epic, Direction string
	Size                    float64
}

type JournalPos struct {
	DealID, Epic string
}

type Classified struct {
	Pos   BrokerPos
	Class string
}

func Reconcile(broker []BrokerPos, journal []JournalPos) (out []Classified, unknown int) {
	jset := map[string]bool{}
	for _, j := range journal {
		if j.DealID != "" {
			jset[j.DealID] = true
		}
	}
	for _, p := range broker {
		c := Unknown
		if jset[p.DealID] {
			c = Managed
		} else if strings.TrimSpace(p.DealID) != "" {
			c = Recovered
		}
		if c == Unknown {
			unknown++
		}
		out = append(out, Classified{Pos: p, Class: c})
	}
	return out, unknown
}

func BlockNewOrders(unknown int) bool { return unknown > 0 }
