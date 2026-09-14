package capsched

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestBoundedConcurrencyAndBackoff(t *testing.T) {
	s := New(2, time.Millisecond)
	var peak, cur atomic.Int32
	ctx := context.Background()
	done := make(chan struct{}, 8)
	for i := 0; i < 8; i++ {
		go func() {
			_ = s.Do(ctx, func(context.Context) error {
				n := cur.Add(1)
				if n > peak.Load() {
					peak.Store(n)
				}
				time.Sleep(5 * time.Millisecond)
				cur.Add(-1)
				return nil
			})
			done <- struct{}{}
		}()
	}
	for i := 0; i < 8; i++ {
		<-done
	}
	if peak.Load() > 2 {
		t.Fatalf("peak=%d", peak.Load())
	}
	if s.Stats().Requests != 8 {
		t.Fatal(s.Stats())
	}
}

func Test429Counted(t *testing.T) {
	s := New(1, time.Millisecond)
	_ = s.Do(context.Background(), func(context.Context) error { return errors.New("429 too many") })
	if s.Stats().Limited != 1 || s.Stats().Errors != 1 {
		t.Fatal(s.Stats())
	}
}
