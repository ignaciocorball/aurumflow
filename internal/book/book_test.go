package book

import "testing"

func TestSnapshotAndDelta(t *testing.T) {
	b := New()
	b.ApplySnapshot(10, []Level{{100, 2}, {99, 1}}, []Level{{101, 3}, {102, 1}})
	bid, ask, ok := b.BestBidAsk()
	if !ok || bid != 100 || ask != 101 {
		t.Fatalf("%v %v %v", bid, ask, ok)
	}
	if err := b.ApplyFuturesDelta(11, 12, 10, []Level{{100, 0}, {99.5, 4}}, nil); err != nil {
		t.Fatal(err)
	}
	bid, _, _ = b.BestBidAsk()
	if bid != 99.5 {
		t.Fatalf("bid=%v", bid)
	}
}

func TestGapResync(t *testing.T) {
	b := New()
	b.ApplySnapshot(10, []Level{{1, 1}}, []Level{{2, 1}})
	if err := b.ApplyFuturesDelta(20, 21, 19, nil, nil); err == nil || b.Synced {
		t.Fatal("expected gap")
	}
	if b.Resyncs != 1 {
		t.Fatalf("resyncs=%d", b.Resyncs)
	}
}

func TestImbalanceAndMicroprice(t *testing.T) {
	b := New()
	b.ApplySnapshot(1, []Level{{10, 8}}, []Level{{11, 2}})
	if imb := b.Imbalance(5); imb < 0.5 {
		t.Fatalf("imb=%v", imb)
	}
	mp, ok := b.Microprice()
	if !ok || mp <= 10 || mp >= 11 {
		t.Fatalf("mp=%v", mp)
	}
}
