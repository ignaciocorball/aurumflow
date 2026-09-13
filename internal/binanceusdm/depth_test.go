package binanceusdm

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"aurumflow/internal/book"
	"aurumflow/internal/md"
)

func TestIsDepthUpdateRejectsNoise(t *testing.T) {
	if isDepthUpdate(streamWrap{}, depthEvent{}) {
		t.Fatal("empty")
	}
	if !isDepthUpdate(streamWrap{Stream: "btcusdt@depth@100ms"}, depthEvent{FirstID: 1, FinalID: 2}) {
		t.Fatal("stream")
	}
	if !isDepthUpdate(streamWrap{}, depthEvent{Event: "depthUpdate", FirstID: 9, FinalID: 10}) {
		t.Fatal("event")
	}
}

func TestHandleMessageIncrementalDelta(t *testing.T) {
	a := New()
	a.Book.ApplySnapshot(10, []book.Level{{Price: 100, Qty: 2}}, []book.Level{{Price: 101, Qty: 3}})
	bus := md.NewBus(8)
	raw := []byte(`{"e":"depthUpdate","E":1,"U":11,"u":12,"pu":10,"b":[["100","0"],["99","4"]],"a":[["101","1"]]}`)
	if err := a.handleMessage(context.Background(), bus, raw, true); err != nil {
		t.Fatal(err)
	}
	select {
	case ev := <-bus.C():
		if ev.Kind != md.KindBookDelta {
			t.Fatalf("kind %s", ev.Kind)
		}
		d, ok := ev.Payload.(md.BookDelta)
		if !ok || d.FinalID != 12 || len(d.Bids) != 2 {
			t.Fatalf("%+v", ev.Payload)
		}
	case <-time.After(time.Second):
		t.Fatal("no delta")
	}
	if _, ok := a.Book.Bids[100]; ok {
		t.Fatal("qty=0 must remove")
	}
	if a.Book.Bids[99] != 4 || a.Book.Asks[101] != 1 {
		t.Fatalf("%+v %+v", a.Book.Bids, a.Book.Asks)
	}
}

func TestForcedGapTriggersResyncPath(t *testing.T) {
	a := New()
	a.Book.ApplySnapshot(10, []book.Level{{Price: 100, Qty: 1}}, []book.Level{{Price: 101, Qty: 1}})
	err := a.Book.ApplyFuturesDelta(20, 21, 19, nil, nil)
	if err == nil {
		t.Fatal("gap must error")
	}
	if a.Book.Synced {
		t.Fatal("must unsync")
	}
	if a.Book.Gaps == 0 {
		t.Fatal("gap counted")
	}
}

func TestCombinedWrapDepthParse(t *testing.T) {
	inner := depthEvent{Event: "depthUpdate", EventTime: 5, FirstID: 2, FinalID: 3, PrevFinal: 1, Bids: [][]string{{"1", "1"}}}
	rawInner, _ := json.Marshal(inner)
	wrap, _ := json.Marshal(streamWrap{Stream: "btcusdt@depth@100ms", Data: rawInner})
	var got streamWrap
	if json.Unmarshal(wrap, &got) != nil || !isDepthUpdate(got, inner) {
		t.Fatal("wrap")
	}
}

func TestCrashDiscardRequiresFreshSnapshot(t *testing.T) {
	a := New()
	a.Book.ApplySnapshot(5, []book.Level{{Price: 100, Qty: 1}}, []book.Level{{Price: 101, Qty: 1}})
	a.Book.Discard()
	if a.Book.Synced {
		t.Fatal("stale book must not remain current truth")
	}
	if err := a.Book.ApplyFuturesDelta(6, 6, 5, nil, nil); err == nil {
		t.Fatal("delta before snapshot")
	}
}
