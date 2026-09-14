package md

import (
	"context"
	"testing"
	"time"
)

func TestClassifyLossBounded(t *testing.T) {
	if ClassifyLoss(ConsumerUI, KindTrade) != LossDisplay {
		t.Fatal("ui")
	}
	if ClassifyLoss(ConsumerBook, KindBookDelta) != LossLossless {
		t.Fatal("book depth")
	}
	if ClassifyLoss(ConsumerBook, KindTrade) != LossTolerant {
		t.Fatal("book trades do not break seq")
	}
	if ClassifyLoss(ConsumerBus, KindTrade) != LossLossless {
		t.Fatal("bus")
	}
}

func TestBusAttributionAndHighWater(t *testing.T) {
	b := NewBus(1)
	ctx := context.Background()
	e := Event{Kind: KindTrade, Provider: "binance", EventTime: time.Now()}
	if !b.Publish(ctx, e) {
		t.Fatal("first")
	}
	if b.Publish(ctx, e) {
		t.Fatal("second")
	}
	tel := b.Telemetry()
	if tel.Drops != 1 || tel.ByKind["trade"] != 1 || tel.ByProv["binance"] != 1 || tel.ByReason[ReasonBackpressure] != 1 {
		t.Fatalf("%+v", tel)
	}
	if tel.Capacity != 1 || tel.HighWater < 1 {
		t.Fatalf("hwm %+v", tel)
	}
}

func TestBusNoFutureLeak(t *testing.T) {
	b := NewBus(2)
	ctx := context.Background()
	t0 := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	b.Publish(ctx, Event{Kind: KindTrade, EventTime: t0, Provider: "binance"})
	ev := <-b.C()
	if ev.EventTime.After(t0) {
		t.Fatal("future")
	}
}
