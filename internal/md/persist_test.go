package md

import (
	"context"
	"testing"
	"time"
)

func TestPersistOverflowAttributed(t *testing.T) {
	q := NewPersistQueue(1)
	e := Event{Kind: KindTrade, EventTime: time.Now(), Provider: "binance"}
	if !q.TryEnqueue(e) {
		t.Fatal("first")
	}
	if q.TryEnqueue(e) {
		t.Fatal("second should drop")
	}
	tel := q.Telemetry()
	if tel.Drops != 1 || tel.Capacity != 1 || tel.HighWater < 1 || tel.LossClass != LossLossless {
		t.Fatalf("%+v", tel)
	}
	if ClassifyLoss(ConsumerRecorder, KindTrade) != LossLossless {
		t.Fatal("recorder must be lossless")
	}
}

func TestBusRatesDistinguishBurst(t *testing.T) {
	b := NewBus(8)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		b.Publish(ctx, Event{Kind: KindTrade, Provider: "binance", EventTime: time.Now()})
	}
	tel := b.Telemetry()
	if tel.Sent != 3 || tel.Depth != 3 || tel.HighWater < 2 {
		t.Fatalf("%+v", tel)
	}
	<-b.C()
	if b.Depth() != 2 {
		t.Fatalf("depth after deq %d", b.Depth())
	}
}
