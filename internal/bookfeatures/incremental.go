package bookfeatures

import (
	"time"

	"aurumflow/internal/book"
)

const (
	VisibleRemoval      = "VISIBLE_LIQUIDITY_REMOVAL"
	VisibleReplenish    = "VISIBLE_LIQUIDITY_REPLENISHMENT"
	KindSnapshot        = "snapshot"
	KindDelta           = "delta"
)

type levelState struct {
	qty           float64
	seen          time.Time
	lastRemoval   time.Time
	lastRemovalQty float64
	refills       int
	latencySum    time.Duration
}

// Incremental is MBP-only. Disappearance may be cancel or trade; we do not claim which.
type Incremental struct {
	AttributionLimit string
	Bid              SideInc
	Ask              SideInc
}

type SideInc struct {
	ReplenishedQty     float64
	ReplenishmentRate  float64
	RefillCount        int
	RefillLatency      float64
	RefillPersistence  float64
	DepletedQty        float64
	DepletionRate      float64
	PersistentDepletion float64
}

func (e *Engine) OnSnapshot(b *book.Book, now time.Time) (Depth, Liquidity) {
	d := Observe(b, now)
	e.seedTracks(b, now)
	e.last = d
	e.lastKind = KindSnapshot
	return d, Liquidity{Attribution: "SNAPSHOT_SEED_NOT_DELTA_FLOW"}
}

func (e *Engine) OnDelta(b *book.Book, now time.Time, bids, asks []book.Level) (Depth, Liquidity) {
	if e.bidTrack == nil {
		e.bidTrack = map[float64]*levelState{}
	}
	if e.askTrack == nil {
		e.askTrack = map[float64]*levelState{}
	}
	var liq Liquidity
	liq.Attribution = "MBP_VISIBLE_ONLY"
	inc := Incremental{AttributionLimit: "CANNOT_SEPARATE_CANCEL_VS_EXECUTION"}
	inc.Bid = e.applySide(e.bidTrack, bids, now, true)
	inc.Ask = e.applySide(e.askTrack, asks, now, false)
	d := Observe(b, now)
	liq.BidDepletion = inc.Bid.DepletedQty
	liq.AskDepletion = inc.Ask.DepletedQty
	liq.BidReplenishment = inc.Bid.ReplenishedQty
	liq.AskReplenishment = inc.Ask.ReplenishedQty
	liq.BidReplCount = inc.Bid.RefillCount
	liq.AskReplCount = inc.Ask.RefillCount
	liq.BidPersistence = persistStates(e.bidTrack, now)
	liq.AskPersistence = persistStates(e.askTrack, now)
	liq.BidRefillLatency = inc.Bid.RefillLatency
	liq.AskRefillLatency = inc.Ask.RefillLatency
	liq.BidReplRate = inc.Bid.ReplenishmentRate
	liq.AskReplRate = inc.Ask.ReplenishmentRate
	liq.BidPersistDepl = inc.Bid.PersistentDepletion
	liq.AskPersistDepl = inc.Ask.PersistentDepletion
	liq.BidRefillPersist = inc.Bid.RefillPersistence
	liq.AskRefillPersist = inc.Ask.RefillPersistence
	e.last = d
	e.lastKind = KindDelta
	e.lastInc = inc
	return d, liq
}

func (e *Engine) LastIncremental() Incremental { return e.lastInc }

func (e *Engine) seedTracks(b *book.Book, now time.Time) {
	e.bidTrack = map[float64]*levelState{}
	e.askTrack = map[float64]*levelState{}
	if b == nil {
		return
	}
	bids, asks := b.CopyTop(0)
	for p, q := range bids {
		e.bidTrack[p] = &levelState{qty: q, seen: now}
	}
	for p, q := range asks {
		e.askTrack[p] = &levelState{qty: q, seen: now}
	}
}

func (e *Engine) applySide(track map[float64]*levelState, levels []book.Level, now time.Time, _ bool) SideInc {
	var s SideInc
	dt := 1.0
	if !e.lastObs.IsZero() {
		if sec := now.Sub(e.lastObs).Seconds(); sec > 0 {
			dt = sec
		}
	}
	e.lastObs = now
	for _, l := range levels {
		st := track[l.Price]
		if st == nil {
			st = &levelState{seen: now}
			track[l.Price] = st
		}
		old := st.qty
		newQty := l.Qty
		if newQty < old {
			removed := old - newQty
			s.DepletedQty += removed
			st.lastRemoval = now
			st.lastRemovalQty = removed
			if newQty == 0 {
				st.seen = time.Time{}
			}
		} else if newQty > old {
			added := newQty - old
			s.ReplenishedQty += added
			s.RefillCount++
			if !st.lastRemoval.IsZero() {
				lat := now.Sub(st.lastRemoval).Seconds()
				s.RefillLatency += lat
				st.latencySum += time.Duration(lat * float64(time.Second))
				st.refills++
				if lat < 2 {
					s.RefillPersistence += added
				}
			}
			if st.seen.IsZero() {
				st.seen = now
			}
		}
		st.qty = newQty
		if newQty == 0 {
			delete(track, l.Price)
		}
	}
	if s.RefillCount > 1 {
		s.RefillLatency /= float64(s.RefillCount)
	}
	if dt > 0 {
		s.ReplenishmentRate = s.ReplenishedQty / dt
		s.DepletionRate = s.DepletedQty / dt
	}
	var persistDepl float64
	for _, st := range track {
		if !st.lastRemoval.IsZero() && st.qty < st.lastRemovalQty && now.Sub(st.lastRemoval) > 500*time.Millisecond {
			persistDepl += st.lastRemovalQty - st.qty
			if persistDepl < 0 {
				persistDepl = 0
			}
		}
	}
	s.PersistentDepletion = persistDepl
	return s
}

func persistStates(track map[float64]*levelState, now time.Time) float64 {
	if len(track) == 0 {
		return 0
	}
	var sum float64
	n := 0
	for _, st := range track {
		if !st.seen.IsZero() {
			sum += now.Sub(st.seen).Seconds()
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}
