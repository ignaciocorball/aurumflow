package microflow

import "time"

// MECHANISM_EVIDENCE_V1 — not FLOW_EXHAUSTION_V2. Does not affect V1 membership.
const FeatureVersion = "MECHANISM_EVIDENCE_V1"

var DefaultWindows = []time.Duration{
	time.Second,
	5 * time.Second,
	15 * time.Second,
	30 * time.Second,
	time.Minute,
	3 * time.Minute,
	5 * time.Minute,
}

type Print struct {
	T          time.Time
	Price, Qty float64
	BuyerMaker bool
}

type WindowStats struct {
	Window           time.Duration `json:"window"`
	TradeCount       int           `json:"trade_count"`
	TradesPerSec     float64       `json:"trades_per_sec"`
	BuyQty           float64       `json:"aggressive_buy_qty"`
	SellQty          float64       `json:"aggressive_sell_qty"`
	BuyNotional      float64       `json:"aggressive_buy_notional"`
	SellNotional     float64       `json:"aggressive_sell_notional"`
	TotalQty         float64       `json:"total_qty"`
	TotalNotional    float64       `json:"total_notional"`
	SignedQty        float64       `json:"signed_qty"`
	SignedNotional   float64       `json:"signed_notional"`
	Imbalance        float64       `json:"buy_sell_imbalance"`
	CVDDelta         float64       `json:"cvd_delta"`
	NotionalVelocity float64       `json:"notional_velocity"`
	QtyVelocity      float64       `json:"quantity_velocity"`
}

type Snapshot struct {
	At      time.Time               `json:"at"`
	Windows map[string]WindowStats  `json:"windows"`
}

type Engine struct {
	prints []Print
}

func New() *Engine { return &Engine{} }

func WindowKey(d time.Duration) string {
	switch d {
	case time.Second:
		return "1s"
	case 5 * time.Second:
		return "5s"
	case 15 * time.Second:
		return "15s"
	case 30 * time.Second:
		return "30s"
	case time.Minute:
		return "1m"
	case 3 * time.Minute:
		return "3m"
	case 5 * time.Minute:
		return "5m"
	default:
		return d.String()
	}
}

func (e *Engine) OnTrade(t time.Time, price, qty float64, buyerMaker bool) {
	if qty <= 0 || price <= 0 || t.IsZero() {
		return
	}
	e.prints = append(e.prints, Print{T: t, Price: price, Qty: qty, BuyerMaker: buyerMaker})
	e.trim(t.Add(-5*time.Minute - time.Second))
}

func (e *Engine) trim(cut time.Time) {
	j := 0
	for j < len(e.prints) && e.prints[j].T.Before(cut) {
		j++
	}
	if j > 0 {
		e.prints = e.prints[j:]
	}
}

// Snapshot uses only prints with EventTime < t0 (no lookahead).
func (e *Engine) Snapshot(t0 time.Time) Snapshot {
	s := Snapshot{At: t0, Windows: map[string]WindowStats{}}
	for _, w := range DefaultWindows {
		s.Windows[WindowKey(w)] = e.window(t0, w)
	}
	return s
}

func (e *Engine) window(t0 time.Time, win time.Duration) WindowStats {
	start := t0.Add(-win)
	var st WindowStats
	st.Window = win
	for _, p := range e.prints {
		if p.T.Before(start) || !p.T.Before(t0) {
			continue
		}
		n := p.Price * p.Qty
		st.TradeCount++
		st.TotalQty += p.Qty
		st.TotalNotional += n
		if p.BuyerMaker {
			st.SellQty += p.Qty
			st.SellNotional += n
			st.SignedQty -= p.Qty
			st.SignedNotional -= n
		} else {
			st.BuyQty += p.Qty
			st.BuyNotional += n
			st.SignedQty += p.Qty
			st.SignedNotional += n
		}
	}
	st.CVDDelta = st.SignedQty
	sec := win.Seconds()
	if sec > 0 {
		st.TradesPerSec = float64(st.TradeCount) / sec
		st.NotionalVelocity = st.TotalNotional / sec
		st.QtyVelocity = st.TotalQty / sec
	}
	if st.TotalQty > 0 {
		st.Imbalance = st.SignedQty / st.TotalQty
	}
	return st
}
