package radar

import (
	"math"
	"time"

	"aurumflow/internal/book"
	"aurumflow/internal/flow"
)

type Engine struct {
	Mode       string
	Instrument string
	Caps       uint32
	Book       *book.Book
	Flow       *flow.Engine
	Liq        book.LiquidityHeuristics
	Hyst       Hysteresis
	Last       PressureSnapshot
	Samples    int
	Events     int64
	LastPrice      float64
	PrevCVD        float64
	PrevCVD2       float64
	cvdAtSnap      float64
	signedWindow   []float64
	minuteBuy      float64
	minuteSell     float64
	minuteN        int
}

func NewEngine(instrument string) *Engine {
	return &Engine{
		Mode: ModeShadow, Instrument: instrument,
		Caps: CapTradeFlow | CapBook,
		Book: book.New(), Flow: &flow.Engine{},
		Hyst: Hysteresis{MinHold: 5 * time.Second},
	}
}

func NewTradeFlowEngine(instrument string) *Engine {
	e := NewEngine(instrument)
	e.Caps = CapTradeFlow
	e.Book = nil
	return e
}

func (e *Engine) OnTrade(price, qty float64, buyerMaker bool) {
	if e.Flow == nil {
		e.Flow = &flow.Engine{}
	}
	prev := e.Flow.CVD()
	e.Flow.OnTrade(flow.Trade{Price: price, Qty: qty, BuyerMaker: buyerMaker})
	e.PrevCVD2 = e.PrevCVD
	e.PrevCVD = prev
	e.LastPrice = price
	e.Events++
	if buyerMaker {
		e.minuteSell += qty
	} else {
		e.minuteBuy += qty
	}
	e.minuteN++
}

func (e *Engine) Has(c uint32) bool { return e.Caps&c != 0 }

func (e *Engine) Snapshot(now time.Time, feedOK bool, typicalQty float64) PressureSnapshot {
	e.Samples++
	tradeOnly := e.Has(CapTradeFlow) && !e.Has(CapBook)
	synced := e.Book != nil && e.Book.Synced && e.Has(CapBook)
	imb5, imb10, imb20 := 0.0, 0.0, 0.0
	if synced {
		imb5 = e.Book.Imbalance(5)
		imb10 = e.Book.Imbalance(10)
		imb20 = e.Book.Imbalance(20)
	}
	bid, ask, okBA := 0.0, 0.0, false
	if e.Book != nil {
		bid, ask, okBA = e.Book.BestBidAsk()
	}
	bidQty, askQty := 0.0, 0.0
	if okBA {
		bidQty, askQty, _ = e.Book.TopQty()
	}
	flags := e.Liq.Observe(bid, ask, bidQty, askQty, typicalQty)
	signed := 0.0
	if e.Flow != nil {
		signed = e.Flow.CVD()
	}
	absorp := 0.0
	if ok, _ := flow.PotentialAbsorption(math.Abs(signed)*0.01, math.Abs(ask-bid), math.Max(bidQty, askQty), typicalQty); ok {
		if signed < 0 && flags.ReplenishBid {
			absorp = 24
		} else if signed > 0 && flags.ReplenishAsk {
			absorp = -24
		}
	}
	flowScore := clamp(signed*0.0001, -25, 25)
	imbScore := clamp((imb5+imb10+imb20)/3*30, -25, 25)
	liqScore := 0.0
	if flags.DepletionAsk && flags.ReplenishBid {
		liqScore = 10
	} else if flags.DepletionBid && flags.ReplenishAsk {
		liqScore = -10
	}
	persist := 0.0
	if flags.PersistentBid {
		persist += 7
	}
	if flags.PersistentAsk {
		persist -= 7
	}
	vol := 0.0
	if okBA {
		spread := ask - bid
		if spread > 0 {
			vol = clamp(5-spread, -5, 5)
		}
	}
	if tradeOnly {
		absorp = 0
		imbScore = 0
		liqScore = 0
		delta := signed - e.cvdAtSnap
		e.cvdAtSnap = signed
		e.signedWindow = append(e.signedWindow, delta)
		if len(e.signedWindow) > 240 {
			e.signedWindow = e.signedWindow[len(e.signedWindow)-240:]
		}
		z := flow.ZScore(delta, mean(e.signedWindow), std(e.signedWindow))
		flowScore = clamp(z*8, -25, 25)
		persist = clamp(z*3, -10, 10)
		if len(e.signedWindow) >= 3 {
			acc := e.signedWindow[len(e.signedWindow)-1] - e.signedWindow[len(e.signedWindow)-2]
			vol = clamp(flow.ZScore(acc, 0, std(e.signedWindow))*2, -5, 5)
		}
	}
	conf := confidence(feedOK, synced || tradeOnly, e.Samples, e.Book)
	s := Compose(PressureSnapshot{
		Instrument: e.Instrument, Timestamp: now,
		AggressiveFlowScore: flowScore, AbsorptionScore: absorp,
		BookImbalanceScore: imbScore, LiquidityScore: liqScore,
		PersistenceScore: persist, VolatilityContext: vol,
		BookSynced: synced, TradeFlowOnly: tradeOnly, Caps: e.Caps,
		Confidence: conf,
	})
	if !e.Hyst.Allow(s.State, now) {
		s.State = e.Hyst.Last
	}
	s.CVD = signed
	s.AggBuy = e.minuteBuy
	s.AggSell = e.minuteSell
	s.TradeVel = float64(e.minuteN)
	e.minuteBuy, e.minuteSell, e.minuteN = 0, 0, 0
	e.Last = s
	return s
}

func confidence(feedOK, synced bool, samples int, b *book.Book) float64 {
	c := 40.0
	if feedOK {
		c += 20
	}
	if synced {
		c += 20
	}
	if samples >= 20 {
		c += 10
	} else if samples >= 5 {
		c += 5
	}
	if b != nil && b.Age() < 2*time.Second {
		c += 10
	}
	if c > 100 {
		c = 100
	}
	return c
}

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := 0.0
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

func std(xs []float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	m := mean(xs)
	ss := 0.0
	for _, x := range xs {
		d := x - m
		ss += d * d
	}
	return math.Sqrt(ss / float64(len(xs)-1))
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
