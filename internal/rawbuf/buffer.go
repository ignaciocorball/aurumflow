package rawbuf

import (
	"sync"
	"time"

	"aurumflow/internal/md"
)

// Buffer keeps a bounded rolling window of raw events (default 120s).
type Buffer struct {
	mu      sync.Mutex
	horizon time.Duration
	maxN    int
	trades  []md.Event
	deltas  []md.Event
	snaps   []md.Event
}

func New() *Buffer {
	return &Buffer{horizon: 120 * time.Second, maxN: 20000}
}

func (b *Buffer) Add(ev md.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	cut := ev.ReceiveTime
	if cut.IsZero() {
		cut = ev.EventTime
	}
	switch ev.Kind {
	case md.KindTrade:
		b.trades = append(b.trades, ev)
	case md.KindBookDelta:
		b.deltas = append(b.deltas, ev)
	case md.KindBookSnapshot:
		b.snaps = append(b.snaps, ev)
	default:
		return
	}
	b.trim(cut)
}

func (b *Buffer) trim(now time.Time) {
	from := now.Add(-b.horizon)
	b.trades = trimEv(b.trades, from, b.maxN)
	b.deltas = trimEv(b.deltas, from, b.maxN)
	b.snaps = trimEv(b.snaps, from, b.maxN)
}

func trimEv(xs []md.Event, from time.Time, maxN int) []md.Event {
	i := 0
	for i < len(xs) {
		t := xs[i].EventTime
		if t.IsZero() {
			t = xs[i].ReceiveTime
		}
		if !t.Before(from) {
			break
		}
		i++
	}
	if i > 0 {
		xs = xs[i:]
	}
	if maxN > 0 && len(xs) > maxN {
		xs = xs[len(xs)-maxN:]
	}
	return xs
}

type Window struct {
	From, To time.Time
	Trades   []md.Event
	Deltas   []md.Event
	Snaps    []md.Event
}

func (b *Buffer) Window(t0 time.Time, before, after time.Duration) Window {
	b.mu.Lock()
	defer b.mu.Unlock()
	from, to := t0.Add(-before), t0.Add(after)
	return Window{
		From: from, To: to,
		Trades: filter(b.trades, from, to),
		Deltas: filter(b.deltas, from, to),
		Snaps:  filter(b.snaps, from, to),
	}
}

func filter(xs []md.Event, from, to time.Time) []md.Event {
	var out []md.Event
	for _, ev := range xs {
		t := ev.EventTime
		if t.IsZero() {
			t = ev.ReceiveTime
		}
		if t.Before(from) || t.After(to) {
			continue
		}
		out = append(out, ev)
	}
	return out
}

func (b *Buffer) Counts() (trades, deltas, snaps int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.trades), len(b.deltas), len(b.snaps)
}
