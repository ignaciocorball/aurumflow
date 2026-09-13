package cftc

import (
	"math"
	"sort"
	"strings"
	"time"
)

const (
	SilverContract = "SILVER - COMMODITY EXCHANGE INC."
	CrudeContract  = "CRUDE OIL, LIGHT SWEET - NEW YORK MERCANTILE EXCHANGE"
	ESContract     = "E-MINI S&P 500 STOCK INDEX - CHICAGO MERCANTILE EXCHANGE"
	NQContract     = "NASDAQ-100 STOCK INDEX (MINI) - CHICAGO MERCANTILE EXCHANGE"
)

var ExactMarkets = map[string]string{
	GoldContract:   "GOLD",
	SilverContract: "SILVER",
	CrudeContract:  "CRUDE_OIL",
	ESContract:     "ES",
	NQContract:     "NQ",
}

func CanonicalMarket(name string) (string, bool) {
	n := strings.TrimSpace(name)
	if id, ok := ExactMarkets[n]; ok {
		return id, true
	}
	if MatchGold(n) {
		return "GOLD", true
	}
	return "", false
}

func matchMarket(rowMarket, want string) bool {
	id, ok := CanonicalMarket(rowMarket)
	if !ok {
		return false
	}
	return id == want || strings.EqualFold(rowMarket, want)
}

func FeaturesMarket(hist []Row, market string, at time.Time) (Row, bool) {
	var usable []Row
	for _, r := range hist {
		if !matchMarket(r.Market, market) {
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
	return usable[len(usable)-1], true
}

func ContextFor(hist []Row, market string, at time.Time) Context {
	cur, ok := FeaturesMarket(hist, market, at)
	if !ok {
		return Context{}
	}
	prevNet := 0.0
	var nets []float64
	for _, r := range hist {
		if !matchMarket(r.Market, market) {
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
		Pct26: percentileMarket(hist, market, at, 26), Pct52: percentileMarket(hist, market, at, 52),
		ZScore: z, Crowding: crowd, Present: true,
	}
}

func percentileMarket(hist []Row, market string, at time.Time, weeks int) float64 {
	var xs []float64
	for _, r := range hist {
		if !matchMarket(r.Market, market) {
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

type PositioningStates struct {
	ManagedMoney Context
	Dealer       Context
	Producer     Context
	AssetManager Context
}

func StatesFromContext(c Context) PositioningStates {
	if !c.Present {
		return PositioningStates{}
	}
	mm := c
	dl := c
	dl.MMNet, dl.WeeklyDelta = c.SDNet, 0
	pr := c
	pr.MMNet, pr.WeeklyDelta = c.PMNet, 0
	am := c
	am.MMNet, am.WeeklyDelta = c.ORNet, 0
	return PositioningStates{ManagedMoney: mm, Dealer: dl, Producer: pr, AssetManager: am}
}
