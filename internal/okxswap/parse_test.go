package okxswap

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"aurumflow/internal/book"
	"aurumflow/internal/md"
)

func TestSubscribePayload(t *testing.T) {
	sub := map[string]any{
		"op": "subscribe",
		"args": []map[string]string{
			{"channel": "books", "instId": InstID},
			{"channel": "trades", "instId": InstID},
		},
	}
	raw, err := json.Marshal(sub)
	if err != nil || !json.Valid(raw) {
		t.Fatal(err)
	}
	if InstID != "BTC-USDT-SWAP" {
		t.Fatal(InstID)
	}
	if Relation != "CORRELATED_PROXY" {
		t.Fatal(Relation)
	}
}

func TestSnapshotAndDeltaParse(t *testing.T) {
	a := New()
	bus := md.NewBus(8)
	ctx := context.Background()
	a.handle(ctx, []byte(`{"arg":{"channel":"books"},"action":"snapshot","data":[{"bids":[["100","2"]],"asks":[["101","3"]],"seqId":10,"prevSeqId":0,"ts":"1"}]}`), bus)
	if !a.Book.Synced || a.Book.LastID != 10 {
		t.Fatal("snapshot")
	}
	select {
	case ev := <-bus.C():
		if ev.Kind != md.KindBookSnapshot {
			t.Fatalf("%s", ev.Kind)
		}
	case <-time.After(time.Second):
		t.Fatal("snap pub")
	}
	a.handle(ctx, []byte(`{"arg":{"channel":"books"},"action":"update","data":[{"bids":[["100","0"],["99","4"]],"asks":[["101","1"]],"seqId":11,"prevSeqId":10,"ts":"2"}]}`), bus)
	if a.Book.LastID != 11 || a.Book.Bids[100] != 0 && a.Book.Bids[99] != 4 {
		if _, ok := a.Book.Bids[100]; ok {
			t.Fatal("qty0")
		}
	}
	select {
	case ev := <-bus.C():
		if ev.Kind != md.KindBookDelta {
			t.Fatalf("%s", ev.Kind)
		}
		d := ev.Payload.(md.BookDelta)
		if d.FirstID != 10 || d.FinalID != 11 {
			t.Fatalf("%+v", d)
		}
	case <-time.After(time.Second):
		t.Fatal("delta pub")
	}
}

func TestSeqGapResync(t *testing.T) {
	b := book.New()
	b.ApplySnapshot(10, []book.Level{{Price: 1, Qty: 1}}, []book.Level{{Price: 2, Qty: 1}})
	if err := b.ApplySeqDelta(40, 39, nil, nil); err == nil || b.Synced {
		t.Fatal("gap")
	}
	b.Discard()
	b.ApplySnapshot(50, []book.Level{{Price: 1, Qty: 1}}, []book.Level{{Price: 2, Qty: 1}})
	if !b.Synced {
		t.Fatal("resync")
	}
}

func TestTradeParse(t *testing.T) {
	a := New()
	bus := md.NewBus(4)
	a.handle(context.Background(), []byte(`{"arg":{"channel":"trades"},"data":[{"px":"100","sz":"0.5","side":"buy","ts":"3"}]}`), bus)
	select {
	case ev := <-bus.C():
		tr := ev.Payload.(md.Trade)
		if tr.Price != 100 || tr.Qty != 0.5 || tr.BuyerMaker {
			t.Fatalf("%+v", tr)
		}
	case <-time.After(time.Second):
		t.Fatal("trade")
	}
}
