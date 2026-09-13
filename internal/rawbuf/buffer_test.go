package rawbuf

import (
	"testing"
	"time"

	"aurumflow/internal/md"
)

func TestBoundAndWindow(t *testing.T) {
	b := New()
	t0 := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	b.Add(md.Event{Kind: md.KindTrade, EventTime: t0.Add(-3 * time.Minute), ReceiveTime: t0})
	b.Add(md.Event{Kind: md.KindTrade, EventTime: t0.Add(-10 * time.Second), ReceiveTime: t0})
	b.Add(md.Event{Kind: md.KindBookDelta, EventTime: t0.Add(-5 * time.Second), ReceiveTime: t0})
	tr, d, _ := b.Counts()
	if tr != 1 || d != 1 {
		t.Fatalf("trimmed %d %d", tr, d)
	}
	w := b.Window(t0, time.Minute, time.Minute)
	if len(w.Trades) != 1 || len(w.Deltas) != 1 {
		t.Fatalf("%+v", w)
	}
}
