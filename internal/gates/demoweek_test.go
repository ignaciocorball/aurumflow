package gates

import (
	"testing"

	"aurumflow/config"
	"aurumflow/internal/money"
)

func okInput() Input {
	return Input{
		Environment: "demo", Host: config.DemoAPIHost, ExecutionMode: string(config.ExecutionDemo),
		AccountOK: true, MarketStatus: "TRADEABLE",
		Spec: money.MonetaryInstrumentSpec{Epic: "GOLD", MinDealSize: 0.01, MaxDealSize: 10, SizeIncrement: 0.01, MoneyPerPriceUnit: 1, ValidationStatus: money.RuntimeValidated},
		JournalOK: true,
	}
}

func TestDemoWeekGates(t *testing.T) {
	if err := DemoWeekTradeAllowed(okInput()); err != nil {
		t.Fatal(err)
	}
	bad := okInput()
	bad.Host = config.LiveAPIHost
	if err := DemoWeekTradeAllowed(bad); err == nil {
		t.Fatal("live host")
	}
	bad = okInput()
	bad.MarketStatus = "CLOSED"
	if err := DemoWeekTradeAllowed(bad); err == nil {
		t.Fatal("closed")
	}
	bad = okInput()
	bad.Spec.MoneyPerPriceUnit = 0
	if err := DemoWeekTradeAllowed(bad); err == nil {
		t.Fatal("monetary")
	}
	bad = okInput()
	bad.KillSwitch = true
	if err := DemoWeekTradeAllowed(bad); err == nil {
		t.Fatal("ks")
	}
}
