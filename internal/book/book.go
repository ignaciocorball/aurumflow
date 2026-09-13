package book

import (
	"fmt"
	"sort"
	"time"
)

type Level struct {
	Price, Qty float64
}

type Book struct {
	Bids, Asks map[float64]float64
	LastID     int64
	Synced     bool
	Resyncs    int
	Updated    time.Time
}

func New() *Book {
	return &Book{Bids: map[float64]float64{}, Asks: map[float64]float64{}}
}

func (b *Book) ApplySnapshot(lastID int64, bids, asks []Level) {
	b.Bids = map[float64]float64{}
	b.Asks = map[float64]float64{}
	for _, l := range bids {
		if l.Qty > 0 {
			b.Bids[l.Price] = l.Qty
		}
	}
	for _, l := range asks {
		if l.Qty > 0 {
			b.Asks[l.Price] = l.Qty
		}
	}
	b.LastID = lastID
	b.Synced = true
	b.Updated = time.Now().UTC()
}

// ApplyFuturesDelta applies a Binance USD-M depth diff.
// First applicable event: U <= lastID+1 <= u. Subsequent: pu == lastID.
func (b *Book) ApplyFuturesDelta(firstID, finalID, prevFinal int64, bids, asks []Level) error {
	if !b.Synced {
		return fmt.Errorf("book not synced")
	}
	if finalID < b.LastID {
		return nil
	}
	if firstID <= b.LastID+1 && b.LastID+1 <= finalID {
		// first after snapshot
	} else if prevFinal != 0 && prevFinal != b.LastID {
		b.Synced = false
		b.Resyncs++
		return fmt.Errorf("gap: pu=%d last=%d", prevFinal, b.LastID)
	} else if prevFinal == 0 && firstID != b.LastID+1 {
		b.Synced = false
		b.Resyncs++
		return fmt.Errorf("gap: U=%d last=%d", firstID, b.LastID)
	}
	for _, l := range bids {
		if l.Qty == 0 {
			delete(b.Bids, l.Price)
		} else {
			b.Bids[l.Price] = l.Qty
		}
	}
	for _, l := range asks {
		if l.Qty == 0 {
			delete(b.Asks, l.Price)
		} else {
			b.Asks[l.Price] = l.Qty
		}
	}
	b.LastID = finalID
	b.Updated = time.Now().UTC()
	return nil
}

func (b *Book) BestBidAsk() (bid, ask float64, ok bool) {
	if b == nil || !b.Synced || len(b.Bids) == 0 || len(b.Asks) == 0 {
		return 0, 0, false
	}
	for p := range b.Bids {
		if bid == 0 || p > bid {
			bid = p
		}
	}
	ask = 0
	for p := range b.Asks {
		if ask == 0 || p < ask {
			ask = p
		}
	}
	return bid, ask, bid > 0 && ask > 0
}

func (b *Book) SpreadMid() (spread, mid float64, ok bool) {
	bid, ask, ok := b.BestBidAsk()
	if !ok {
		return 0, 0, false
	}
	return ask - bid, (ask + bid) / 2, true
}

func (b *Book) Microprice() (float64, bool) {
	bid, ask, ok := b.BestBidAsk()
	if !ok {
		return 0, false
	}
	bq, aq := b.Bids[bid], b.Asks[ask]
	den := bq + aq
	if den <= 0 {
		return (bid + ask) / 2, true
	}
	return (ask*bq + bid*aq) / den, true
}

func (b *Book) Imbalance(n int) float64 {
	bids := topN(b.Bids, n, true)
	asks := topN(b.Asks, n, false)
	var bv, av float64
	for _, l := range bids {
		bv += l.Qty
	}
	for _, l := range asks {
		av += l.Qty
	}
	if bv+av == 0 {
		return 0
	}
	return (bv - av) / (bv + av)
}

func topN(m map[float64]float64, n int, highFirst bool) []Level {
	out := make([]Level, 0, len(m))
	for p, q := range m {
		out = append(out, Level{Price: p, Qty: q})
	}
	sort.Slice(out, func(i, j int) bool {
		if highFirst {
			return out[i].Price > out[j].Price
		}
		return out[i].Price < out[j].Price
	})
	if n > 0 && len(out) > n {
		out = out[:n]
	}
	return out
}

func (b *Book) Age() time.Duration {
	if b.Updated.IsZero() {
		return 0
	}
	return time.Since(b.Updated)
}
