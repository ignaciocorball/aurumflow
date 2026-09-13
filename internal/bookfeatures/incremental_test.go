package bookfeatures

import (
	"testing"
	"time"

	"aurumflow/internal/book"
	"aurumflow/internal/md"
)

func TestIncrementalReplenishDeplete(t *testing.T) {
	e := New()
	b := book.New()
	now := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	b.ApplySnapshot(1, []book.Level{{Price: 100, Qty: 10}}, []book.Level{{Price: 101, Qty: 8}})
	e.OnSnapshot(b, now)
	_, liq := e.OnDelta(b, now.Add(100*time.Millisecond), []book.Level{{Price: 100, Qty: 2}}, nil)
	if liq.BidDepletion < 8 {
		t.Fatalf("removal %+v", liq)
	}
	if liq.Attribution == "" {
		t.Fatal("must label visible-only")
	}
	_, liq2 := e.OnDelta(b, now.Add(200*time.Millisecond), []book.Level{{Price: 100, Qty: 9}}, nil)
	if liq2.BidReplenishment < 6 || liq2.BidReplCount < 1 {
		t.Fatalf("refill %+v", liq2)
	}
	if liq2.BidRefillLatency <= 0 {
		t.Fatalf("latency %+v", liq2)
	}
	inc := e.LastIncremental()
	if inc.AttributionLimit == "" {
		t.Fatal("cancel vs execution must remain unknown")
	}
}

func TestSnapshotIsNotDeltaFlow(t *testing.T) {
	e := New()
	b := book.New()
	now := time.Now().UTC()
	b.ApplySnapshot(1, []book.Level{{Price: 1, Qty: 1}}, []book.Level{{Price: 2, Qty: 1}})
	_, liq := e.OnSnapshot(b, now)
	if liq.BidReplenishment != 0 || liq.Attribution == "" {
		t.Fatalf("%+v", liq)
	}
}

func TestCouplingWindows(t *testing.T) {
	e := New()
	t0 := time.Date(2026, 9, 13, 15, 0, 0, 0, time.UTC)
	d := Depth{Available: true, Mid: 100, Microprice: 100.01, Bid5: 5, Ask5: 4}
	stats := e.OnTradeCouple(t0, 2, true, d, Liquidity{AskDepletion: 1, BidReplenishment: 0.5})
	if len(stats) != 6 {
		t.Fatalf("windows %d", len(stats))
	}
	if stats[0].Window != (100 * time.Millisecond).String() || stats[5].Window != (5 * time.Second).String() {
		t.Fatalf("%+v", stats)
	}
}

func TestLiveReplayParity(t *testing.T) {
	now := time.Date(2026, 9, 13, 16, 0, 0, 0, time.UTC)
	live := book.New()
	eng := New()
	snap := md.Event{Kind: md.KindBookSnapshot, EventTime: now, Payload: md.BookDelta{
		FinalID: 10, Bids: []md.Level{{Price: 100, Qty: 5}}, Asks: []md.Level{{Price: 101, Qty: 5}},
	}}
	live.ApplySnapshot(10, []book.Level{{Price: 100, Qty: 5}}, []book.Level{{Price: 101, Qty: 5}})
	eng.OnSnapshot(live, now)
	delta := md.Event{Kind: md.KindBookDelta, EventTime: now.Add(time.Millisecond), Payload: md.BookDelta{
		FirstID: 11, FinalID: 11, PrevFinal: 10,
		Bids: []md.Level{{Price: 100, Qty: 3}}, Asks: []md.Level{{Price: 101, Qty: 6}},
	}}
	_ = live.ApplyFuturesDelta(11, 11, 10, []book.Level{{Price: 100, Qty: 3}}, []book.Level{{Price: 101, Qty: 6}})
	liveD, _ := eng.OnDelta(live, now.Add(time.Millisecond), []book.Level{{Price: 100, Qty: 3}}, []book.Level{{Price: 101, Qty: 6}})
	rep := Replay([]md.Event{snap, delta})
	if !FeaturesClose(liveD, rep.Depth, 1e-9) {
		t.Fatalf("live %+v replay %+v", liveD, rep.Depth)
	}
	if !rep.Book.Synced || rep.Deltas != 1 {
		t.Fatal("replay book")
	}
}
