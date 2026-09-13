package sessions

import (
	"testing"
	"time"
)

func TestPhaseAndHandoff(t *testing.T) {
	s := DefaultSchedule()
	asia := time.Date(2026, 9, 14, 2, 0, 0, 0, time.UTC)
	eu := time.Date(2026, 9, 14, 9, 0, 0, 0, time.UTC)
	us := time.Date(2026, 9, 14, 15, 0, 0, 0, time.UTC)
	if PhaseAt(asia, s) != Asia {
		t.Fatal(PhaseAt(asia, s))
	}
	if PhaseAt(eu, s) != Europe && PhaseAt(eu, s) != Overlap {
		t.Fatal(PhaseAt(eu, s))
	}
	if PhaseAt(us, s) != US && PhaseAt(us, s) != Overlap {
		t.Fatal(PhaseAt(us, s))
	}
	h, ok := Handoff(Asia, Europe, eu)
	if !ok || h.Event == "" {
		t.Fatal(h, ok)
	}
	if MarketHours("US100", "CLOSED", us) != MarketClosed {
		t.Fatal("broker status must win")
	}
	if MarketHours("US100", "TRADEABLE", us) != MarketOpen {
		t.Fatal("tradeable")
	}
	if MarketHours("BTC", "", asia) != MarketOpen {
		t.Fatal("btc")
	}
}
