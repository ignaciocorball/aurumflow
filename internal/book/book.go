package book

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type Level struct {
	Price, Qty float64
}

type Book struct {
	mu         sync.RWMutex
	Bids, Asks map[float64]float64
	LastID     int64
	Synced     bool
	Resyncs    int
	Gaps       int
	Updated    time.Time
	LastResyncReason string
	resyncLog  []string
}

func New() *Book {
	return &Book{Bids: map[float64]float64{}, Asks: map[float64]float64{}}
}

func (b *Book) Discard() {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Bids = map[float64]float64{}
	b.Asks = map[float64]float64{}
	b.LastID = 0
	b.Synced = false
	b.Updated = time.Time{}
}

func (b *Book) ApplySnapshot(lastID int64, bids, asks []Level) {
	b.mu.Lock()
	defer b.mu.Unlock()
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

func (b *Book) ApplyFuturesDelta(firstID, finalID, prevFinal int64, bids, asks []Level) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.Synced {
		return fmt.Errorf("book not synced")
	}
	if finalID < b.LastID {
		return nil
	}
	if firstID <= b.LastID+1 && b.LastID+1 <= finalID {
		// first after snapshot
	} else if prevFinal != 0 && prevFinal != b.LastID {
		b.noteResyncLocked("sequence_mismatch")
		b.Synced = false
		b.Gaps++
		return fmt.Errorf("gap: pu=%d last=%d", prevFinal, b.LastID)
	} else if prevFinal == 0 && firstID != b.LastID+1 {
		b.noteResyncLocked("sequence_mismatch")
		b.Synced = false
		b.Gaps++
		return fmt.Errorf("gap: U=%d last=%d", firstID, b.LastID)
	}
	applyLevels(b.Bids, bids)
	applyLevels(b.Asks, asks)
	b.LastID = finalID
	b.Updated = time.Now().UTC()
	return nil
}

func applyLevels(side map[float64]float64, levels []Level) {
	for _, l := range levels {
		if l.Qty == 0 {
			delete(side, l.Price)
		} else {
			side[l.Price] = l.Qty
		}
	}
}

func (b *Book) BestBidAsk() (bid, ask float64, ok bool) {
	if b == nil {
		return 0, 0, false
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.bestBidAskLocked()
}

func (b *Book) bestBidAskLocked() (bid, ask float64, ok bool) {
	if !b.Synced || len(b.Bids) == 0 || len(b.Asks) == 0 {
		return 0, 0, false
	}
	for p := range b.Bids {
		if bid == 0 || p > bid {
			bid = p
		}
	}
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
	if b == nil {
		return 0, false
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	bid, ask, ok := b.bestBidAskLocked()
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
	if b == nil {
		return 0
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
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
	if b == nil {
		return 0
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.Updated.IsZero() {
		return 0
	}
	return time.Since(b.Updated)
}

func (b *Book) AgeAt(now time.Time) time.Duration {
	if b == nil {
		return 0
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.Updated.IsZero() {
		return 0
	}
	return now.Sub(b.Updated)
}

func (b *Book) FeaturesAccepted() bool {
	if b == nil {
		return false
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Synced
}

func (b *Book) DepthQty(n int) (bidQty, askQty float64) {
	if b == nil {
		return 0, 0
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	if !b.Synced {
		return 0, 0
	}
	for _, l := range topN(b.Bids, n, true) {
		bidQty += l.Qty
	}
	for _, l := range topN(b.Asks, n, false) {
		askQty += l.Qty
	}
	return bidQty, askQty
}

func (b *Book) ApplySeqDelta(seq, prev int64, bids, asks []Level) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.Synced {
		return fmt.Errorf("book not synced")
	}
	if prev != 0 && prev != b.LastID {
		b.noteResyncLocked("sequence_mismatch")
		b.Synced = false
		b.Gaps++
		return fmt.Errorf("gap: prev=%d last=%d", prev, b.LastID)
	}
	if prev == 0 && seq != b.LastID+1 && b.LastID != 0 {
		b.noteResyncLocked("sequence_mismatch")
		b.Synced = false
		b.Gaps++
		return fmt.Errorf("gap: seq=%d last=%d", seq, b.LastID)
	}
	applyLevels(b.Bids, bids)
	applyLevels(b.Asks, asks)
	b.LastID = seq
	b.Updated = time.Now().UTC()
	return nil
}

// CopyTop returns copies of the top n bid/ask levels for feature engines.
func (b *Book) CopyTop(n int) (bids, asks map[float64]float64) {
	if b == nil {
		return map[float64]float64{}, map[float64]float64{}
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	if n <= 0 {
		return copyTopMap(b.Bids, len(b.Bids), true), copyTopMap(b.Asks, len(b.Asks), false)
	}
	return copyTopMap(b.Bids, n, true), copyTopMap(b.Asks, n, false)
}

func (b *Book) MarkUnsynced() {
	if b == nil {
		return
	}
	b.mu.Lock()
	b.Synced = false
	b.mu.Unlock()
}

func (b *Book) IncResyncs() {
	b.NoteResync("unspecified")
}

func (b *Book) NoteResync(reason string) {
	if b == nil {
		return
	}
	b.mu.Lock()
	b.noteResyncLocked(reason)
	b.mu.Unlock()
}

func (b *Book) noteResyncLocked(reason string) {
	if reason == "" {
		reason = "unspecified"
	}
	b.Resyncs++
	b.LastResyncReason = reason
	if len(b.resyncLog) < 8 {
		b.resyncLog = append(b.resyncLog, reason)
	}
}

func (b *Book) Meta() (synced bool, lastID int64, gaps, resyncs int) {
	if b == nil {
		return false, 0, 0, 0
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Synced, b.LastID, b.Gaps, b.Resyncs
}

func (b *Book) ResyncReason() string {
	if b == nil {
		return ""
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.LastResyncReason
}

func (b *Book) ResyncLog() string {
	if b == nil {
		return ""
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := ""
	for i, r := range b.resyncLog {
		if i > 0 {
			out += ","
		}
		out += r
	}
	return out
}

func copyTopMap(m map[float64]float64, n int, highFirst bool) map[float64]float64 {
	out := map[float64]float64{}
	for _, l := range topN(m, n, highFirst) {
		out[l.Price] = l.Qty
	}
	return out
}
