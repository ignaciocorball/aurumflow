package gates

import (
	"fmt"
	"strings"

	"aurumflow/config"
	"aurumflow/internal/money"
)

type Input struct {
	Environment    string
	Host           string
	ExecutionMode  string
	AccountOK      bool
	KillSwitch     bool
	MarketStatus   string
	Spec           money.MonetaryInstrumentSpec
	JournalOK      bool
	DailyDDBlocked bool
	DataStale      bool
	UnknownPositions int
}

func DemoWeekTradeAllowed(in Input) error {
	if strings.EqualFold(in.Environment, "live") || config.IsLiveCapitalHost(in.Host) {
		return fmt.Errorf("LIVE detected")
	}
	if !config.IsDemoCapitalHost(in.Host) {
		return fmt.Errorf("wrong host")
	}
	if !strings.EqualFold(in.Environment, "demo") {
		return fmt.Errorf("environment must be demo")
	}
	if in.ExecutionMode != string(config.ExecutionDemo) {
		return fmt.Errorf("execution mode must be DEMO")
	}
	if !in.AccountOK {
		return fmt.Errorf("wrong account")
	}
	if in.KillSwitch {
		return fmt.Errorf("kill switch active")
	}
	if !strings.EqualFold(in.MarketStatus, "TRADEABLE") {
		return fmt.Errorf("GOLD CLOSED or not TRADEABLE")
	}
	if !in.Spec.DealingComplete() {
		return fmt.Errorf("InstrumentSpec invalid")
	}
	if in.Spec.MoneyPerPriceUnit <= 0 || in.Spec.ValidationStatus != money.RuntimeValidated {
		return fmt.Errorf("monetary spec unvalidated")
	}
	if !in.JournalOK {
		return fmt.Errorf("journal failure")
	}
	if in.DailyDDBlocked {
		return fmt.Errorf("daily DD blocked")
	}
	if in.DataStale {
		return fmt.Errorf("stale market data")
	}
	if in.UnknownPositions > 0 {
		return fmt.Errorf("unknown recovered positions block new orders")
	}
	return nil
}
