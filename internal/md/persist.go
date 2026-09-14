package md

import "sync/atomic"

// PersistQueue is a bounded lossless-intent disk path. Overflow is attributed
// separately from the live bus so book/micro consumers are not coupled to IO.
type PersistQueue struct {
	ch      chan Event
	cap     int
	dropped atomic.Int64
	enq     atomic.Int64
	deq     atomic.Int64
	hwm     atomic.Int64
}

func NewPersistQueue(n int) *PersistQueue {
	if n < 1 {
		n = 1024
	}
	return &PersistQueue{ch: make(chan Event, n), cap: n}
}

func (q *PersistQueue) TryEnqueue(e Event) bool {
	if q == nil {
		return false
	}
	if d := int64(len(q.ch)); d > q.hwm.Load() {
		q.hwm.Store(d)
	}
	select {
	case q.ch <- e:
		q.enq.Add(1)
		return true
	default:
		q.dropped.Add(1)
		return false
	}
}

func (q *PersistQueue) C() <-chan Event {
	if q == nil {
		return nil
	}
	return q.ch
}

func (q *PersistQueue) MarkDeq() {
	if q != nil {
		q.deq.Add(1)
	}
}

func (q *PersistQueue) Telemetry() Telemetry {
	t := Telemetry{Consumer: ConsumerRecorder, LossClass: LossLossless}
	if q == nil {
		return t
	}
	t.Capacity, t.Depth, t.HighWater = q.cap, len(q.ch), q.hwm.Load()
	t.Sent, t.Drops = q.enq.Load(), q.dropped.Load()
	t.ByReason = map[string]int64{ReasonBackpressure: t.Drops}
	return t
}
