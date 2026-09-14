package capsched

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var ErrRateLimited = errors.New("capital read rate limited")

type Stats struct {
	Requests int64
	Errors   int64
	Limited  int64
	P50Ms    float64
	P95Ms    float64
	PerMin   float64
}

type Scheduler struct {
	conc    int
	minGap  time.Duration
	sem     chan struct{}
	mu      sync.Mutex
	last    time.Time
	lats    []float64
	reqs    atomic.Int64
	errs    atomic.Int64
	lim     atomic.Int64
	started time.Time
}

func New(conc int, minGap time.Duration) *Scheduler {
	if conc < 1 {
		conc = 1
	}
	if conc > 8 {
		conc = 8
	}
	if minGap <= 0 {
		minGap = 250 * time.Millisecond
	}
	return &Scheduler{conc: conc, minGap: minGap, sem: make(chan struct{}, conc), started: time.Now()}
}

func (s *Scheduler) Concurrency() int { return s.conc }

func (s *Scheduler) Do(ctx context.Context, fn func(context.Context) error) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.sem <- struct{}{}:
	}
	defer func() { <-s.sem }()
	s.pace(ctx)
	t0 := time.Now()
	err := fn(ctx)
	dt := time.Since(t0).Seconds() * 1000
	s.reqs.Add(1)
	s.mu.Lock()
	s.lats = append(s.lats, dt)
	if len(s.lats) > 2048 {
		s.lats = s.lats[len(s.lats)-1024:]
	}
	s.mu.Unlock()
	if err != nil {
		s.errs.Add(1)
		if is429(err) {
			s.lim.Add(1)
			select {
			case <-ctx.Done():
			case <-time.After(time.Second):
			}
		}
	}
	return err
}

func (s *Scheduler) pace(ctx context.Context) {
	s.mu.Lock()
	wait := s.minGap - time.Since(s.last)
	if wait < 0 {
		wait = 0
	}
	s.last = time.Now().Add(wait)
	s.mu.Unlock()
	if wait > 0 {
		select {
		case <-ctx.Done():
		case <-time.After(wait):
		}
	}
}

func (s *Scheduler) Stats() Stats {
	st := Stats{Requests: s.reqs.Load(), Errors: s.errs.Load(), Limited: s.lim.Load()}
	if up := time.Since(s.started).Minutes(); up > 0 {
		st.PerMin = float64(st.Requests) / up
	}
	s.mu.Lock()
	st.P50Ms, st.P95Ms = pct(s.lats, 0.5), pct(s.lats, 0.95)
	s.mu.Unlock()
	return st
}

func is429(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrRateLimited) {
		return true
	}
	s := err.Error()
	return len(s) >= 3 && (s[:3] == "429" || contains429(s))
}

func contains429(s string) bool {
	for i := 0; i+2 < len(s); i++ {
		if s[i] == '4' && s[i+1] == '2' && s[i+2] == '9' {
			return true
		}
	}
	return false
}

func pct(xs []float64, p float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	cp := append([]float64(nil), xs...)
	for i := 1; i < len(cp); i++ {
		j := i
		for j > 0 && cp[j] < cp[j-1] {
			cp[j], cp[j-1] = cp[j-1], cp[j]
			j--
		}
	}
	idx := int(float64(len(cp)-1) * p)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(cp) {
		idx = len(cp) - 1
	}
	return cp[idx]
}
