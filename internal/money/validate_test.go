package money

import "testing"

func TestUPLMatchAndDirectedPnL(t *testing.T) {
	s := MonetaryInstrumentSpec{MoneyPerPriceUnit: 1}
	pnl, err := ExpectedPnLDirected(s, 0.0001, 77299.40, 77249.40, "BUY")
	if err != nil || pnl > -0.004 || pnl < -0.006 {
		t.Fatalf("pnl=%v err=%v", pnl, err)
	}
	sell, err := ExpectedPnLDirected(s, 0.0001, 77299.40, 77249.40, "SELL")
	if err != nil || sell < 0.004 {
		t.Fatalf("sell=%v", sell)
	}
	if !UPLMatches(-0.005, 0, 0.01, 0.25) {
		t.Fatal("tiny UPL vs 0 should tolerate")
	}
	if RuntimeOK(1, 3) || !RuntimeOK(2, 3) {
		t.Fatal("runtime threshold")
	}
}
