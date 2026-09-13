package md

import (
	"context"
	"testing"
	"time"
)

func TestBus_BackpressureDrops(t *testing.T) {
	b := NewBus(1)
	ctx := context.Background()
	e := Event{Kind: KindTrade, EventTime: time.Now(), Provider: "t", Venue: "v", Instrument: "X"}
	if !b.Publish(ctx, e) {
		t.Fatal("first")
	}
	if b.Publish(ctx, e) {
		t.Fatal("second should drop")
	}
	sent, drop := b.Stats()
	if sent != 1 || drop != 1 {
		t.Fatalf("sent=%d drop=%d", sent, drop)
	}
}

func TestCaps(t *testing.T) {
	c := CapTrades | CapBookDelta
	if !c.Has(CapTrades) || c.Has(CapCandles) {
		t.Fatal(c)
	}
}
