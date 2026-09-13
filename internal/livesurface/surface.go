package livesurface

import (
	"math"
	"strings"
	"time"
)

const Version = "MARKET_FEATURES_V1"

const (
	MomStrongUp   = "STRONG_UP"
	MomUp         = "UP"
	MomFlat       = "FLAT"
	MomDown       = "DOWN"
	MomStrongDown = "STRONG_DOWN"
	MomUnknown    = "UNKNOWN"
	MomClosed     = "CLOSED"
	MomStale      = "STALE"

	VolLow     = "LOW"
	VolNormal  = "NORMAL"
	VolHigh    = "HIGH"
	VolExtreme = "EXTREME"
	VolUnknown = "UNKNOWN"

	BreadthStrongBroad = "STRONG_BROAD"
	BreadthBroad       = "BROAD"
	BreadthNarrow      = "NARROW"
	BreadthNegative    = "NEGATIVE"
	BreadthUnknown     = "UNKNOWN"

	SessPreopen = "PREOPEN"
	SessOpen    = "OPEN"
	SessClosed  = "CLOSED"
	SessUnknown = "UNKNOWN"

	RelCorrContext = "CORRELATION_CONTEXT"
)

type Quote struct {
	Market       string
	Epic         string
	Bid          float64
	Ask          float64
	Mid          float64
	Spread       float64
	MarketStatus string
	EventTime    time.Time
	ReceivedAt   time.Time
	Age          time.Duration
	Stale        bool
	Live         bool
}

type Candle struct {
	Time  time.Time
	Close float64
	High  float64
	Low   float64
}

type Features struct {
	Market     string
	Ret5m      *float64
	Ret15m     *float64
	Ret1h      *float64
	Ret4h      *float64
	Ret1d      *float64
	Vol15m     *float64
	Vol1h      *float64
	Vol4h      *float64
	Vol1d      *float64
	Range      *float64
	SpreadNorm *float64
	SessionHL  *float64
	Momentum   string
	VolState   string
	RelStrength string
	EventTime  time.Time
	ReceivedAt time.Time
	Age        time.Duration
	Live       bool
	Closed     bool
	Stale      bool
}

type Breadth struct {
	State          string
	Advancing      int
	Declining      int
	MedianMegaRet  *float64
	US100vsUS500   *float64
	Explanation    string
	Mega           map[string]*float64
}

type Precious struct {
	GoldRS, SilverRS string
	Ratio            *float64
	RatioChange      *float64
	VolDivergence    string
}

type Energy struct {
	OilMomentum    string
	PhysicalContext string
}

type Europe struct {
	BreadthProxy string
	Leadership   string
	Session      string
}

type Asia struct {
	Leadership string
	Session    string
}

type Edge struct {
	A, B     string
	Kind     string
	Corr1h   *float64
	Corr4h   *float64
	Corr1d   *float64
}

type Frame struct {
	AsOf     time.Time
	Quotes   map[string]Quote
	Features map[string]Features
	Breadth  Breadth
	Precious Precious
	Energy   Energy
	Europe   Europe
	Asia     Asia
	Graph    []Edge
}

func NewQuote(market, epic string, bid, ask float64, status string, event, recv time.Time, staleAfter time.Duration) Quote {
	mid := 0.0
	if bid > 0 && ask > 0 {
		mid = (bid + ask) / 2
	} else if bid > 0 {
		mid = bid
	} else {
		mid = ask
	}
	q := Quote{
		Market: market, Epic: epic, Bid: bid, Ask: ask, Mid: mid,
		Spread: math.Max(0, ask-bid), MarketStatus: status,
		EventTime: event.UTC(), ReceivedAt: recv.UTC(),
	}
	if !event.IsZero() {
		q.Age = recv.UTC().Sub(event.UTC())
	} else if !recv.IsZero() {
		q.Age = 0
	}
	q.Stale = staleAfter > 0 && q.Age > staleAfter
	q.Live = LiveForRanking(status, q.Stale)
	return q
}

func LiveForRanking(status string, stale bool) bool {
	if stale {
		return false
	}
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "TRADEABLE", "OPEN", "ON":
		return true
	default:
		return false
	}
}

func StaleAfter(d time.Duration) time.Duration {
	if d <= 0 {
		return 2 * time.Minute
	}
	return d
}

func ptr(v float64) *float64 { return &v }

func retAt(closes []Candle, now time.Time, lookback time.Duration) *float64 {
	if len(closes) < 2 {
		return nil
	}
	latest := closes[len(closes)-1]
	if latest.Close <= 0 {
		return nil
	}
	target := now.Add(-lookback)
	var prev *Candle
	for i := len(closes) - 2; i >= 0; i-- {
		if !closes[i].Time.After(target) && closes[i].Close > 0 {
			c := closes[i]
			prev = &c
			break
		}
	}
	if prev == nil {
		return nil
	}
	v := (latest.Close - prev.Close) / prev.Close
	return &v
}

func realizedVol(closes []Candle, window int) *float64 {
	if window < 2 || len(closes) < window+1 {
		return nil
	}
	rets := make([]float64, 0, window)
	start := len(closes) - window - 1
	for i := start + 1; i < len(closes); i++ {
		a, b := closes[i-1].Close, closes[i].Close
		if a <= 0 || b <= 0 {
			return nil
		}
		rets = append(rets, math.Log(b/a))
	}
	if len(rets) < 2 {
		return nil
	}
	m := 0.0
	for _, r := range rets {
		m += r
	}
	m /= float64(len(rets))
	ss := 0.0
	for _, r := range rets {
		d := r - m
		ss += d * d
	}
	v := math.Sqrt(ss / float64(len(rets)-1))
	return &v
}

func sessionHL(closes []Candle) *float64 {
	if len(closes) < 2 {
		return nil
	}
	latest := closes[len(closes)-1]
	hi, lo := latest.High, latest.Low
	if hi <= 0 || lo <= 0 {
		hi, lo = latest.Close, latest.Close
		for _, c := range closes {
			if c.Close > hi {
				hi = c.Close
			}
			if c.Close > 0 && (lo == 0 || c.Close < lo) {
				lo = c.Close
			}
		}
	} else {
		for _, c := range closes {
			if c.High > hi {
				hi = c.High
			}
			if c.Low > 0 && c.Low < lo {
				lo = c.Low
			}
		}
	}
	if hi <= lo || latest.Close <= 0 {
		return nil
	}
	v := (latest.Close - lo) / (hi - lo)
	return &v
}

func FeaturesFrom(market string, q Quote, hist []Candle, now time.Time) Features {
	f := Features{
		Market: market, EventTime: q.EventTime, ReceivedAt: q.ReceivedAt,
		Age: q.Age, Live: q.Live, Closed: !q.Live && strings.EqualFold(q.MarketStatus, "CLOSED"),
		Stale: q.Stale, Momentum: MomUnknown, VolState: VolUnknown,
	}
	if q.Mid > 0 && q.Spread > 0 {
		sn := q.Spread / q.Mid
		f.SpreadNorm = &sn
	}
	if !q.Live {
		if q.Stale {
			f.Momentum = MomStale
		} else if strings.EqualFold(q.MarketStatus, "CLOSED") {
			f.Momentum = MomClosed
		}
		return f
	}
	if len(hist) == 0 {
		return f
	}
	now = now.UTC()
	f.Ret5m = retAt(hist, now, 5*time.Minute)
	f.Ret15m = retAt(hist, now, 15*time.Minute)
	f.Ret1h = retAt(hist, now, time.Hour)
	f.Ret4h = retAt(hist, now, 4*time.Hour)
	f.Ret1d = retAt(hist, now, 24*time.Hour)
	f.Vol15m = realizedVol(hist, 15)
	f.Vol1h = realizedVol(hist, 60)
	f.Vol4h = realizedVol(hist, 16)
	f.Vol1d = realizedVol(hist, 24)
	f.Range = sessionHL(hist)
	f.SessionHL = f.Range
	f.Momentum = MomentumOf(f.Ret1h, f.Ret4h)
	own := rollingMedian(hist, 20)
	f.VolState = VolatilityOf(f.Vol1h, own)
	return f
}

func rollingMedian(hist []Candle, n int) *float64 {
	v := realizedVol(hist, n)
	return v
}

func MomentumOf(ret1h, ret4h *float64) string {
	if ret1h == nil {
		return MomUnknown
	}
	r1 := *ret1h
	r4 := 0.0
	if ret4h != nil {
		r4 = *ret4h
	}
	switch {
	case r1 > 0.005 && r4 > 0.01:
		return MomStrongUp
	case r1 < -0.005 && r4 < -0.01:
		return MomStrongDown
	case r1 > 0.0015:
		return MomUp
	case r1 < -0.0015:
		return MomDown
	default:
		return MomFlat
	}
}

func VolatilityOf(current, ownHist *float64) string {
	if current == nil || *current < 0 {
		return VolUnknown
	}
	if ownHist == nil || *ownHist <= 0 {
		return VolUnknown
	}
	ratio := *current / *ownHist
	switch {
	case ratio < 0.6:
		return VolLow
	case ratio < 1.5:
		return VolNormal
	case ratio < 2.5:
		return VolHigh
	default:
		return VolExtreme
	}
}

func RelativeStrength(rets map[string]*float64, live map[string]bool) map[string]string {
	out := map[string]string{}
	type pair struct {
		id  string
		ret float64
	}
	var xs []pair
	for id, r := range rets {
		if r == nil || !live[id] {
			out[id] = MomUnknown
			continue
		}
		xs = append(xs, pair{id, *r})
	}
	if len(xs) < 2 {
		for id := range rets {
			if live[id] && rets[id] != nil {
				out[id] = MomFlat
			}
		}
		return out
	}
	mean := 0.0
	for _, x := range xs {
		mean += x.ret
	}
	mean /= float64(len(xs))
	for _, x := range xs {
		d := x.ret - mean
		switch {
		case d > 0.002:
			out[x.id] = MomUp
		case d < -0.002:
			out[x.id] = MomDown
		default:
			out[x.id] = MomFlat
		}
	}
	return out
}

func PeerLeadership(a, b string, rets map[string]*float64, live map[string]bool) string {
	if !live[a] || !live[b] || rets[a] == nil || rets[b] == nil {
		return "UNKNOWN"
	}
	if *rets[a] > *rets[b]+0.0005 {
		return a
	}
	if *rets[b] > *rets[a]+0.0005 {
		return b
	}
	return "TIED"
}

func BreadthOf(us100, us500, us30 *float64, mega map[string]*float64, live bool) Breadth {
	b := Breadth{State: BreadthUnknown, Mega: mega, Explanation: "insufficient live US equity quotes"}
	if !live {
		b.Explanation = "US equity quotes closed or stale — not treated as live breadth"
		return b
	}
	adv, dec := 0, 0
	var vals []float64
	for _, r := range mega {
		if r == nil {
			continue
		}
		vals = append(vals, *r)
		if *r > 0 {
			adv++
		} else if *r < 0 {
			dec++
		}
	}
	b.Advancing, b.Declining = adv, dec
	if len(vals) > 0 {
		med := median(vals)
		b.MedianMegaRet = &med
	}
	if us100 != nil && us500 != nil {
		d := *us100 - *us500
		b.US100vsUS500 = &d
	}
	idxUp := (us100 != nil && *us100 > 0) || (us500 != nil && *us500 > 0) || (us30 != nil && *us30 > 0)
	idxDown := (us100 != nil && *us100 < 0) && (us500 == nil || *us500 < 0)
	switch {
	case idxDown || (us100 != nil && *us100 < 0 && us500 != nil && *us500 < 0):
		b.State = BreadthNegative
		b.Explanation = "US index returns negative"
	case idxUp && adv >= 4:
		b.State = BreadthStrongBroad
		b.Explanation = "indices up and most mega-caps advancing"
	case idxUp && adv >= 3:
		b.State = BreadthBroad
		b.Explanation = "indices up with majority mega-cap participation"
	case idxUp && adv <= 2:
		b.State = BreadthNarrow
		b.Explanation = "indices up but mega-cap participation narrow"
	default:
		b.State = BreadthUnknown
		b.Explanation = "mixed or incomplete mega-cap tape"
	}
	return b
}

func median(xs []float64) float64 {
	ys := append([]float64{}, xs...)
	for i := 0; i < len(ys); i++ {
		for j := i + 1; j < len(ys); j++ {
			if ys[j] < ys[i] {
				ys[i], ys[j] = ys[j], ys[i]
			}
		}
	}
	n := len(ys)
	if n == 0 {
		return 0
	}
	if n%2 == 1 {
		return ys[n/2]
	}
	return (ys[n/2-1] + ys[n/2]) / 2
}

func RollingCorr(a, b []Candle, minN int) *float64 {
	if minN < 3 {
		minN = 3
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if n < minN {
		return nil
	}
	a = a[len(a)-n:]
	b = b[len(b)-n:]
	var xa, xb []float64
	for i := 1; i < n; i++ {
		if a[i-1].Close <= 0 || a[i].Close <= 0 || b[i-1].Close <= 0 || b[i].Close <= 0 {
			continue
		}
		xa = append(xa, math.Log(a[i].Close/a[i-1].Close))
		xb = append(xb, math.Log(b[i].Close/b[i-1].Close))
	}
	if len(xa) < minN-1 {
		return nil
	}
	ma, mb := 0.0, 0.0
	for i := range xa {
		ma += xa[i]
		mb += xb[i]
	}
	ma /= float64(len(xa))
	mb /= float64(len(xb))
	num, da, db := 0.0, 0.0, 0.0
	for i := range xa {
		x, y := xa[i]-ma, xb[i]-mb
		num += x * y
		da += x * x
		db += y * y
	}
	den := math.Sqrt(da * db)
	if den == 0 {
		return nil
	}
	v := num / den
	return &v
}

func DefaultGraph() [][2]string {
	return [][2]string{
		{"US100", "US500"},
		{"US100", "US30"},
		{"GOLD", "SILVER"},
		{"DE40", "UK100"},
		{"J225", "CN50"},
		{"GOLD", "US100"},
		{"OIL_CRUDE", "US500"},
	}
}

func BuildGraph(hist map[string][]Candle, live map[string]bool) []Edge {
	var out []Edge
	for _, p := range DefaultGraph() {
		e := Edge{A: p[0], B: p[1], Kind: RelCorrContext}
		if !live[p[0]] || !live[p[1]] {
			out = append(out, e)
			continue
		}
		ha, hb := hist[p[0]], hist[p[1]]
		e.Corr1h = RollingCorr(tailDur(ha, time.Hour), tailDur(hb, time.Hour), 8)
		e.Corr4h = RollingCorr(tailDur(ha, 4*time.Hour), tailDur(hb, 4*time.Hour), 8)
		e.Corr1d = RollingCorr(ha, hb, 12)
		out = append(out, e)
	}
	return out
}

func tailDur(xs []Candle, d time.Duration) []Candle {
	if len(xs) == 0 {
		return nil
	}
	cut := xs[len(xs)-1].Time.Add(-d)
	var out []Candle
	for _, c := range xs {
		if !c.Time.Before(cut) {
			out = append(out, c)
		}
	}
	return out
}

func Assemble(now time.Time, quotes map[string]Quote, hist map[string][]Candle) Frame {
	f := Frame{AsOf: now.UTC(), Quotes: quotes, Features: map[string]Features{}}
	live := map[string]bool{}
	rets := map[string]*float64{}
	for id, q := range quotes {
		feat := FeaturesFrom(id, q, hist[id], now)
		f.Features[id] = feat
		live[id] = q.Live
		rets[id] = feat.Ret1h
	}
	rs := RelativeStrength(rets, live)
	for id, feat := range f.Features {
		if v, ok := rs[id]; ok {
			feat.RelStrength = v
			f.Features[id] = feat
		}
	}
	usLive := live["US100"] || live["US500"] || live["US30"]
	mega := map[string]*float64{}
	for _, id := range []string{"AAPL", "MSFT", "NVDA", "META", "AMZN"} {
		if ft, ok := f.Features[id]; ok {
			mega[id] = ft.Ret1h
		}
	}
	var u1, u5, u3 *float64
	if ft, ok := f.Features["US100"]; ok {
		u1 = ft.Ret1h
	}
	if ft, ok := f.Features["US500"]; ok {
		u5 = ft.Ret1h
	}
	if ft, ok := f.Features["US30"]; ok {
		u3 = ft.Ret1h
	}
	f.Breadth = BreadthOf(u1, u5, u3, mega, usLive)
	f.Precious = preciousOf(f.Features, live, quotes)
	f.Energy = Energy{OilMomentum: momOrUnknown(f.Features["OIL_CRUDE"]), PhysicalContext: "UNKNOWN"}
	f.Europe = Europe{
		Leadership:   PeerLeadership("DE40", "UK100", rets, live),
		Session:      sessionWord(live["DE40"], live["UK100"]),
		BreadthProxy: europeBreadth(rets, live),
	}
	f.Asia = Asia{
		Leadership: PeerLeadership("J225", "CN50", rets, live),
		Session:    sessionWord(live["J225"], live["CN50"]),
	}
	f.Graph = BuildGraph(hist, live)
	return f
}

func momOrUnknown(f Features) string {
	if f.Momentum == "" {
		return MomUnknown
	}
	return f.Momentum
}

func sessionWord(a, b bool) string {
	if a || b {
		return SessOpen
	}
	return SessClosed
}

func europeBreadth(rets map[string]*float64, live map[string]bool) string {
	if !live["DE40"] || !live["UK100"] || rets["DE40"] == nil || rets["UK100"] == nil {
		return BreadthUnknown
	}
	up := 0
	if *rets["DE40"] > 0 {
		up++
	}
	if *rets["UK100"] > 0 {
		up++
	}
	switch up {
	case 2:
		return BreadthBroad
	case 1:
		return BreadthNarrow
	default:
		return BreadthNegative
	}
}

func preciousOf(feat map[string]Features, live map[string]bool, quotes map[string]Quote) Precious {
	p := Precious{GoldRS: MomUnknown, SilverRS: MomUnknown, VolDivergence: "UNKNOWN"}
	rs := RelativeStrength(map[string]*float64{
		"GOLD":   feat["GOLD"].Ret1h,
		"SILVER": feat["SILVER"].Ret1h,
	}, map[string]bool{"GOLD": live["GOLD"], "SILVER": live["SILVER"]})
	p.GoldRS, p.SilverRS = rs["GOLD"], rs["SILVER"]
	gq, sq := quotes["GOLD"], quotes["SILVER"]
	if gq.Mid > 0 && sq.Mid > 0 {
		r := gq.Mid / sq.Mid
		p.Ratio = &r
	}
	if feat["GOLD"].Ret1d != nil && feat["SILVER"].Ret1d != nil {
		// ratio change ≈ gold 1d minus silver 1d when both live
		if live["GOLD"] && live["SILVER"] {
			d := *feat["GOLD"].Ret1d - *feat["SILVER"].Ret1d
			p.RatioChange = &d
		}
	}
	if feat["GOLD"].VolState != VolUnknown && feat["SILVER"].VolState != VolUnknown && live["GOLD"] && live["SILVER"] {
		if feat["GOLD"].VolState != feat["SILVER"].VolState {
			p.VolDivergence = feat["GOLD"].VolState + "_vs_" + feat["SILVER"].VolState
		} else {
			p.VolDivergence = "ALIGNED"
		}
	}
	return p
}
