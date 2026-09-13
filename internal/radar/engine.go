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
	Book       *book.Book
	Flow       *flow.Engine
	Liq        book.LiquidityHeuristics
	Hyst       Hysteresis
	Last       PressureSnapshot
	Samples    int
	Events     int64
}

func NewEngine(instrument string) *Engine {
	return &Engine{
		Mode: ModeShadow, Instrument: instrument,
		Book: book.New(), Flow: &flow.Engine{},
		Hyst: Hysteresis{MinHold: 5 * time.Second},
	}
}

func (e *Engine) OnTrade(price, qty float64, buyerMaker bool) {
	if e.Flow == nil {
		e.Flow = &flow.Engine{}
	}
	e.Flow.OnTrade(flow.Trade{Price: price, Qty: qty, BuyerMaker: buyerMaker})
	e.Events++
}

func (e *Engine) Snapshot(now time.Time, feedOK bool, typicalQty float64) PressureSnapshot {
	e.Samples++
	synced := e.Book != nil && e.Book.Synced
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
	s := Compose(PressureSnapshot{
		Instrument: e.Instrument, Timestamp: now,
		AggressiveFlowScore: flowScore, AbsorptionScore: absorp,
		BookImbalanceScore: imbScore, LiquidityScore: liqScore,
		PersistenceScore: persist, VolatilityContext: vol,
		BookSynced: synced, Confidence: confidence(feedOK, synced, e.Samples, e.Book),
	})
	if !e.Hyst.Allow(s.State, now) {
		s.State = e.Hyst.Last
	}
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

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
