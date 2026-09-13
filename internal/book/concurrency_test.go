package book

import (
	"sync"
	"testing"
	"time"
)

func TestConcurrentApplyAndRead(t *testing.T) {
	b := New()
	b.ApplySnapshot(1, []Level{{100, 1}}, []Level{{101, 1}})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			_ = b.ApplyFuturesDelta(int64(i+2), int64(i+2), int64(i+1), []Level{{100, float64(i%5 + 1)}}, nil)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			_, _, _ = b.BestBidAsk()
			_, _ = b.CopyTop(5)
			_ = b.Imbalance(5)
		}
	}()
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
}
