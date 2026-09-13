package radar

import (
	"testing"
	"time"

	"aurumflow/internal/book"
)

func TestEngineSnapshotUsesBookSync(t *testing.T) {
	e := NewEngine("BTCUSDT")
	e.Book.ApplySnapshot(1, []book.Level{{Price: 100, Qty: 5}}, []book.Level{{Price: 101, Qty: 5}})
	e.OnTrade(101, 2, false)
	s := e.Snapshot(time.Unix(10, 0).UTC(), true, 1)
	if s.Instrument != "BTCUSDT" || s.Confidence <= 0 {
		t.Fatalf("%+v", s)
	}
	e.Book.Synced = false
	s = e.Snapshot(time.Unix(20, 0).UTC(), true, 1)
	if s.State != StateNoTrade {
		t.Fatalf("unsynced state=%s", s.State)
	}
}
