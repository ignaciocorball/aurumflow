package radar

import (
	"testing"
	"time"
)

func TestTradeFlowDoesNotInventBook(t *testing.T) {
	e := NewTradeFlowEngine("BTCUSDT")
	if e.Has(CapBook) {
		t.Fatal("book cap")
	}
	e.OnTrade(100, 2, false)
	e.OnTrade(100.2, 1, true)
	s := e.Snapshot(time.Unix(10, 0).UTC(), true, 1)
	if s.BookImbalanceScore != 0 || s.LiquidityScore != 0 || s.AbsorptionScore != 0 {
		t.Fatalf("invented book features %+v", s)
	}
	if s.State == "" || !s.TradeFlowOnly {
		t.Fatalf("%+v", s)
	}
}
