package md

import (
	"context"
	"sync"
	"sync/atomic"
)

type Bus struct {
	ch      chan Event
	cap     int
	dropped atomic.Int64
	sent    atomic.Int64
	hwm     atomic.Int64
	mu      sync.Mutex
	byKind  map[Kind]int64
	byProv  map[string]int64
	byReason map[string]int64
}

func NewBus(buf int) *Bus {
	if buf < 1 {
		buf = 64
	}
	return &Bus{ch: make(chan Event, buf), cap: buf, byKind: map[Kind]int64{}, byProv: map[string]int64{}, byReason: map[string]int64{}}
}

func (b *Bus) Capacity() int { return b.cap }

func (b *Bus) Depth() int {
	if b == nil {
		return 0
	}
	return len(b.ch)
}

func (b *Bus) HighWater() int64 {
	if b == nil {
		return 0
	}
	return b.hwm.Load()
}

func (b *Bus) Publish(ctx context.Context, e Event) bool {
	if b == nil {
		return false
	}
	if d := int64(len(b.ch)); d > b.hwm.Load() {
		b.hwm.Store(d)
	}
	select {
	case <-ctx.Done():
		b.noteDrop(e, ReasonCtxCancel)
		return false
	case b.ch <- e:
		b.sent.Add(1)
		return true
	default:
		b.noteDrop(e, ReasonBackpressure)
		return false
	}
}

func (b *Bus) noteDrop(e Event, reason string) {
	b.dropped.Add(1)
	b.mu.Lock()
	b.byKind[e.Kind]++
	prov := e.Provider
	if prov == "" {
		prov = "unknown"
	}
	b.byProv[prov]++
	b.byReason[reason]++
	b.mu.Unlock()
}

func (b *Bus) C() <-chan Event { return b.ch }

func (b *Bus) Stats() (sent, dropped int64) {
	return b.sent.Load(), b.dropped.Load()
}

type Telemetry struct {
	Capacity int              `json:"capacity"`
	Depth    int              `json:"depth"`
	HighWater int64           `json:"high_water"`
	Sent     int64            `json:"sent"`
	Drops    int64            `json:"drops"`
	ByKind   map[string]int64 `json:"drops_by_event_type"`
	ByProv   map[string]int64 `json:"drops_by_provider"`
	ByReason map[string]int64 `json:"drops_by_reason"`
	Consumer string           `json:"consumer"`
	LossClass string          `json:"loss_class"`
}

func (b *Bus) Telemetry() Telemetry {
	t := Telemetry{Consumer: ConsumerBus, LossClass: ClassifyLoss(ConsumerBus, KindTrade)}
	if b == nil {
		return t
	}
	t.Capacity, t.Depth, t.HighWater = b.cap, len(b.ch), b.hwm.Load()
	t.Sent, t.Drops = b.Stats()
	b.mu.Lock()
	t.ByKind = map[string]int64{}
	for k, n := range b.byKind {
		t.ByKind[string(k)] = n
	}
	t.ByProv = copyI64(b.byProv)
	t.ByReason = copyI64(b.byReason)
	b.mu.Unlock()
	return t
}

func copyI64(in map[string]int64) map[string]int64 {
	out := make(map[string]int64, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

type Counter struct {
	mu sync.Mutex
	n  int
}

func (c *Counter) Inc() {
	c.mu.Lock()
	c.n++
	c.mu.Unlock()
}
