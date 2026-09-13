package bookfeatures

import (
	"math"

	"aurumflow/internal/book"
	"aurumflow/internal/md"
)

type ReplayResult struct {
	Book   *book.Book
	Depth  Depth
	Deltas int
	Snaps  int
}

// Replay applies captured live events offline. Live and replay engines must match.
func Replay(events []md.Event) ReplayResult {
	b := book.New()
	eng := New()
	var last Depth
	var deltas, snaps int
	for _, ev := range events {
		d, ok := ev.Payload.(md.BookDelta)
		if !ok {
			if raw, ok2 := ev.Payload.(*md.BookDelta); ok2 && raw != nil {
				d = *raw
				ok = true
			}
		}
		if !ok {
			continue
		}
		bl := toBookLevels(d.Bids)
		al := toBookLevels(d.Asks)
		switch ev.Kind {
		case md.KindBookSnapshot:
			snaps++
			b.ApplySnapshot(d.FinalID, bl, al)
			last, _ = eng.OnSnapshot(b, ev.EventTime)
		case md.KindBookDelta:
			deltas++
			_ = b.ApplyFuturesDelta(d.FirstID, d.FinalID, d.PrevFinal, bl, al)
			if !b.Synced {
				_ = b.ApplySeqDelta(d.FinalID, d.PrevFinal, bl, al)
			}
			last, _ = eng.OnDelta(b, ev.EventTime, bl, al)
		}
	}
	return ReplayResult{Book: b, Depth: last, Deltas: deltas, Snaps: snaps}
}

func toBookLevels(xs []md.Level) []book.Level {
	out := make([]book.Level, len(xs))
	for i, l := range xs {
		out[i] = book.Level{Price: l.Price, Qty: l.Qty}
	}
	return out
}

func FeaturesClose(a, b Depth, eps float64) bool {
	if a.Available != b.Available {
		return false
	}
	return near(a.Spread, b.Spread, eps) && near(a.Mid, b.Mid, eps) &&
		near(a.Microprice, b.Microprice, eps) && near(a.Imb5, b.Imb5, eps)
}

func near(a, b, eps float64) bool {
	return math.Abs(a-b) <= eps
}
