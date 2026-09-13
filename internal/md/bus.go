package md

import (
	"context"
	"sync"
	"sync/atomic"
)

type Bus struct {
	ch      chan Event
	dropped atomic.Int64
	sent    atomic.Int64
}

func NewBus(buf int) *Bus {
	if buf < 1 {
		buf = 64
	}
	return &Bus{ch: make(chan Event, buf)}
}

func (b *Bus) Publish(ctx context.Context, e Event) bool {
	if b == nil {
		return false
	}
	select {
	case <-ctx.Done():
		b.dropped.Add(1)
		return false
	case b.ch <- e:
		b.sent.Add(1)
		return true
	default:
		b.dropped.Add(1)
		return false
	}
}

func (b *Bus) C() <-chan Event { return b.ch }

func (b *Bus) Stats() (sent, dropped int64) {
	return b.sent.Load(), b.dropped.Load()
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
