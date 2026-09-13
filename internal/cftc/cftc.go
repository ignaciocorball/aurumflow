package cftc

import (
	"math"
	"sort"
	"strings"
	"time"
)

// Official GOLD futures identity in CFTC Disaggregated COT.
const GoldContract = "GOLD - COMMODITY EXCHANGE INC."

type Row struct {
	Market     string
	AsOf       time.Time // typically Tuesday
	Available  time.Time // Friday publication (conservative +3d 21:30 UTC)
	MMLong     float64
	MMShort    float64
	MMNet      float64
	SDNet      float64
	PMNet      float64
	ORNet      float64
}

func AvailableAt(asOf time.Time) time.Time {
	// COT as-of Tuesday; released the following Friday ~15:30 ET ≈ 19:30/20:30 UTC. Use Saturday 00:00 UTC conservative.
	return asOf.UTC().AddDate(0, 0, 4)
}

func MatchGold(name string) bool {
	n := strings.ToUpper(strings.TrimSpace(name))
	return n == GoldContract || n == "GOLD - COMMODITY EXCHANGE INC."
}

func Features(hist []Row, at time.Time) (Row, bool) {
	var usable []Row
	for _, r := range hist {
		if !MatchGold(r.Market) {
			continue
		}
		if r.Available.IsZero() {
			r.Available = AvailableAt(r.AsOf)
		}
		if !r.Available.After(at) {
			usable = append(usable, r)
		}
	}
	if len(usable) == 0 {
		return Row{}, false
	}
	sort.Slice(usable, func(i, j int) bool { return usable[i].AsOf.Before(usable[j].AsOf) })
	cur := usable[len(usable)-1]
	return cur, true
}

// Context is a no-lookahead GOLD COT snapshot used as SLOW_CONTEXT only.
type Context struct {
	Row
	WeeklyDelta   float64
	Pct26         float64
	Pct52         float64
	ZScore        float64
	Crowding      float64
	Present       bool
}

func ContextAt(hist []Row, at time.Time) Context {
	cur, ok := Features(hist, at)
	if !ok {
		return Context{}
	}
	prevNet := 0.0
	var nets []float64
	for _, r := range hist {
		if !MatchGold(r.Market) {
			continue
		}
		avail := r.Available
		if avail.IsZero() {
			avail = AvailableAt(r.AsOf)
		}
		if avail.After(at) {
			continue
		}
		if r.AsOf.Before(cur.AsOf) {
			prevNet = r.MMNet
		}
		nets = append(nets, r.MMNet)
	}
	z := 0.0
	if len(nets) >= 2 {
		m := 0.0
		for _, x := range nets {
			m += x
		}
		m /= float64(len(nets))
		ss := 0.0
		for _, x := range nets {
			d := x - m
			ss += d * d
		}
		sd := 0.0
		if len(nets) > 1 {
			sd = math.Sqrt(ss / float64(len(nets)-1))
		}
		if sd > 0 {
			z = (cur.MMNet - m) / sd
		}
	}
	den := cur.MMLong + cur.MMShort
	crowd := 0.0
	if den != 0 {
		crowd = abs(cur.MMNet) / den
	}
	return Context{
		Row: cur, WeeklyDelta: cur.MMNet - prevNet,
		Pct26: PercentileNet(hist, at, 26), Pct52: PercentileNet(hist, at, 52),
		ZScore: z, Crowding: crowd, Present: true,
	}
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func PercentileNet(hist []Row, at time.Time, weeks int) float64 {
	var xs []float64
	for _, r := range hist {
		if !MatchGold(r.Market) {
			continue
		}
		avail := r.Available
		if avail.IsZero() {
			avail = AvailableAt(r.AsOf)
		}
		if avail.After(at) {
			continue
		}
		xs = append(xs, r.MMNet)
	}
	if len(xs) == 0 {
		return 0
	}
	if weeks > 0 && len(xs) > weeks {
		xs = xs[len(xs)-weeks:]
	}
	cur := xs[len(xs)-1]
	n := 0
	for _, x := range xs {
		if x <= cur {
			n++
		}
	}
	return 100 * float64(n) / float64(len(xs))
}
