package bookfeatures

import (
	"math"
	"time"

	"aurumflow/internal/book"
)

type Depth struct {
	Spread, Mid, Microprice float64
	Bid1, Bid5, Bid10, Bid20 float64
	Ask1, Ask5, Ask10, Ask20 float64
	Imb1, Imb5, Imb10, Imb20 float64
	DepthDelta               float64
	Available                bool
}

type Liquidity struct {
	BidDepletion, AskDepletion         float64
	BidReplenishment, AskReplenishment float64
	BidReplCount, AskReplCount         int
	BidPersistence, AskPersistence     float64
	ReplConsumedBid, ReplConsumedAsk   float64
	Sweep                              string
	Attribution                        string
	BidRefillLatency, AskRefillLatency float64
	BidReplRate, AskReplRate           float64
	BidPersistDepl, AskPersistDepl     float64
	BidRefillPersist, AskRefillPersist float64
}

type Sample struct {
	T              time.Time
	Bid1, Ask1     float64
	BidPx, AskPx   float64
	BidTop, AskTop map[float64]float64
}

type Engine struct {
	hist     []Sample
	bidSeen  map[float64]time.Time
	askSeen  map[float64]time.Time
	bidQty   map[float64]float64
	askQty   map[float64]float64
	last     Depth
	lastKind string
	lastObs  time.Time
	lastInc  Incremental
	bidTrack map[float64]*levelState
	askTrack map[float64]*levelState
	couple   Coupler
}

func New() *Engine {
	return &Engine{
		bidSeen: map[float64]time.Time{},
		askSeen: map[float64]time.Time{},
		bidQty:  map[float64]float64{},
		askQty:  map[float64]float64{},
	}
}

func Observe(b *book.Book, now time.Time) Depth {
	var d Depth
	if b == nil || !b.FeaturesAccepted() {
		return d
	}
	bid, ask, ok := b.BestBidAsk()
	if !ok {
		return d
	}
	d.Available = true
	d.Spread = ask - bid
	d.Mid = (ask + bid) / 2
	if mp, mok := b.Microprice(); mok {
		d.Microprice = mp
	}
	d.Bid1, d.Ask1 = b.DepthQty(1)
	d.Bid5, d.Ask5 = b.DepthQty(5)
	d.Bid10, d.Ask10 = b.DepthQty(10)
	d.Bid20, d.Ask20 = b.DepthQty(20)
	d.Imb1 = b.Imbalance(1)
	d.Imb5 = b.Imbalance(5)
	d.Imb10 = b.Imbalance(10)
	d.Imb20 = b.Imbalance(20)
	return d
}

func (e *Engine) OnBook(b *book.Book, now time.Time) (Depth, Liquidity) {
	d := Observe(b, now)
	var liq Liquidity
	if !d.Available {
		return d, liq
	}
	d.DepthDelta = (d.Bid5 + d.Ask5) - (e.last.Bid5 + e.last.Ask5)
	bid, ask, _ := b.BestBidAsk()
	bidTop, askTop := b.CopyTop(10)
	s := Sample{T: now, Bid1: d.Bid1, Ask1: d.Ask1, BidPx: bid, AskPx: ask, BidTop: bidTop, AskTop: askTop}
	if e.last.Available {
		if d.Bid5 < e.last.Bid5 {
			liq.BidDepletion = e.last.Bid5 - d.Bid5
		}
		if d.Ask5 < e.last.Ask5 {
			liq.AskDepletion = e.last.Ask5 - d.Ask5
		}
		liq.BidReplenishment, liq.BidReplCount, liq.ReplConsumedBid = refill(e.bidQty, s.BidTop, e.last.Bid5, d.Bid5)
		liq.AskReplenishment, liq.AskReplCount, liq.ReplConsumedAsk = refill(e.askQty, s.AskTop, e.last.Ask5, d.Ask5)
	}
	for p := range s.BidTop {
		if e.bidSeen[p].IsZero() {
			e.bidSeen[p] = now
		}
	}
	for p := range s.AskTop {
		if e.askSeen[p].IsZero() {
			e.askSeen[p] = now
		}
	}
	liq.BidPersistence = persist(e.bidSeen, s.BidTop, now)
	liq.AskPersistence = persist(e.askSeen, s.AskTop, now)
	e.bidQty = s.BidTop
	e.askQty = s.AskTop
	e.last = d
	e.hist = append(e.hist, s)
	cut := now.Add(-2 * time.Minute)
	i := 0
	for i < len(e.hist) && e.hist[i].T.Before(cut) {
		i++
	}
	if i > 0 {
		e.hist = e.hist[i:]
	}
	return d, liq
}

func copySide(m map[float64]float64, n int, highFirst bool) map[float64]float64 {
	out := map[float64]float64{}
	if m == nil {
		return out
	}
	type kv struct{ p, q float64 }
	xs := make([]kv, 0, len(m))
	for p, q := range m {
		xs = append(xs, kv{p, q})
	}
	for i := 1; i < len(xs); i++ {
		j := i
		for j > 0 {
			less := xs[j].p < xs[j-1].p
			if highFirst {
				less = xs[j].p > xs[j-1].p
			}
			if !less {
				break
			}
			xs[j], xs[j-1] = xs[j-1], xs[j]
			j--
		}
	}
	if n > 0 && len(xs) > n {
		xs = xs[:n]
	}
	for _, x := range xs {
		out[x.p] = x.q
	}
	return out
}

func refill(prev map[float64]float64, now map[float64]float64, prevDepth, nowDepth float64) (qty float64, n int, ratio float64) {
	consumed := prevDepth - nowDepth
	if consumed < 0 {
		consumed = 0
	}
	for p, q := range now {
		old := prev[p]
		if q > old {
			qty += q - old
			n++
		}
	}
	if consumed > 0 {
		ratio = qty / consumed
	}
	return qty, n, ratio
}

func persist(seen map[float64]time.Time, now map[float64]float64, t time.Time) float64 {
	if len(now) == 0 {
		return 0
	}
	var sum float64
	for p := range now {
		if !seen[p].IsZero() {
			sum += t.Sub(seen[p]).Seconds()
		}
	}
	return sum / float64(len(now))
}

func PotentialSweep(tradeNotional float64, levelsCrossed int, typical float64) string {
	if typical <= 0 {
		typical = 1
	}
	if levelsCrossed >= 2 && tradeNotional >= typical*3 {
		return "POTENTIAL_SWEEP"
	}
	return ""
}

func RobustZ(xs []float64, v float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	cp := append([]float64(nil), xs...)
	for i := 1; i < len(cp); i++ {
		j := i
		for j > 0 && cp[j] < cp[j-1] {
			cp[j], cp[j-1] = cp[j-1], cp[j]
			j--
		}
	}
	med := cp[len(cp)/2]
	dev := make([]float64, len(cp))
	for i, x := range cp {
		dev[i] = math.Abs(x - med)
	}
	for i := 1; i < len(dev); i++ {
		j := i
		for j > 0 && dev[j] < dev[j-1] {
			dev[j], dev[j-1] = dev[j-1], dev[j]
			j--
		}
	}
	mad := dev[len(dev)/2]
	if mad <= 1e-12 {
		return 0
	}
	return (v - med) / (1.4826 * mad)
}
