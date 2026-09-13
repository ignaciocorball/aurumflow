package bookfeatures

import "time"

var CouplingWindows = []time.Duration{
	100 * time.Millisecond,
	250 * time.Millisecond,
	500 * time.Millisecond,
	time.Second,
	2 * time.Second,
	5 * time.Second,
}

type Aggression struct {
	T        time.Time
	Qty      float64
	Sell     bool
	Mid      float64
	Micro    float64
	Support  float64
	Opposing float64
}

type WindowStat struct {
	Window                    string  `json:"window"`
	AggressiveQty             float64 `json:"aggressive_qty"`
	SupportingDepthBefore     float64 `json:"supporting_depth_before"`
	SupportingRemoval         float64 `json:"supporting_removal"`
	SupportingRefill          float64 `json:"supporting_refill"`
	OpposingDepletion         float64 `json:"opposing_depletion"`
	MidDisplacement           float64 `json:"mid_displacement"`
	MicropriceDisplacement    float64 `json:"microprice_displacement"`
}

type Coupler struct {
	events []Aggression
}

func (c *Coupler) OnAggression(a Aggression) {
	c.events = append(c.events, a)
	cut := a.T.Add(-6 * time.Second)
	i := 0
	for i < len(c.events) && c.events[i].T.Before(cut) {
		i++
	}
	if i > 0 {
		c.events = c.events[i:]
	}
}

func (c *Coupler) Measure(now time.Time, mid, micro, supportNow, opposeNow, supportRemoval, supportRefill, opposeDepl float64) []WindowStat {
	out := make([]WindowStat, 0, len(CouplingWindows))
	for _, w := range CouplingWindows {
		from := now.Add(-w)
		var qty, mid0, micro0, supp0, opp0 float64
		n := 0
		for _, ev := range c.events {
			if ev.T.Before(from) || ev.T.After(now) {
				continue
			}
			qty += ev.Qty
			if n == 0 {
				mid0, micro0, supp0, opp0 = ev.Mid, ev.Micro, ev.Support, ev.Opposing
			}
			n++
		}
		st := WindowStat{Window: w.String(), AggressiveQty: qty}
		if n > 0 {
			st.SupportingDepthBefore = supp0
			st.SupportingRemoval = supportRemoval
			st.SupportingRefill = supportRefill
			st.OpposingDepletion = opposeDepl
			st.MidDisplacement = mid - mid0
			st.MicropriceDisplacement = micro - micro0
			_ = opp0
		}
		out = append(out, st)
	}
	return out
}

func (e *Engine) OnTradeCouple(t time.Time, qty float64, aggressiveSell bool, d Depth, liq Liquidity) []WindowStat {
	support, oppose := d.Bid5, d.Ask5
	if aggressiveSell {
		support, oppose = d.Bid5, d.Ask5
	} else {
		support, oppose = d.Ask5, d.Bid5
	}
	e.couple.OnAggression(Aggression{
		T: t, Qty: qty, Sell: aggressiveSell,
		Mid: d.Mid, Micro: d.Microprice, Support: support, Opposing: oppose,
	})
	return e.couple.Measure(t, d.Mid, d.Microprice, support, oppose, liq.BidDepletion+liq.AskDepletion, liq.BidReplenishment+liq.AskReplenishment, liq.AskDepletion+liq.BidDepletion)
}
