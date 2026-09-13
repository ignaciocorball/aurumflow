package stratrade

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func WriteReport(dir string) error {
	if strings.TrimSpace(dir) == "" {
		return nil
	}
	a, err := Reconstruct(filepath.Dir(dir), filepath.Base(dir))
	if err != nil {
		a, err = Reconstruct(dir, filepath.Base(dir))
	}
	if err != nil {
		pre := filepath.Join(dir, "pre_signal.json")
		if _, stat := os.Stat(pre); stat != nil {
			return err
		}
		a, _ = Reconstruct(filepath.Dir(dir), filepath.Base(dir))
	}
	var b strings.Builder
	b.WriteString("# First strategy trade forensic report\n\n")
	b.WriteString("No post-hoc strategy judgement. Evidence only.\n\n")
	if a.PreSignal != nil {
		fmt.Fprintf(&b, "## Why did Legacy enter?\n\n")
		fmt.Fprintf(&b, "Origin `%s` score `%d` session `%s` direction `%s` evidence `%s`.\n\n",
			a.PreSignal.Origin, a.PreSignal.LegacyScore, a.PreSignal.Session, a.PreSignal.Direction, a.PreSignal.LegacyEvidence)
		fmt.Fprintf(&b, "## What did risk permit?\n\n")
		fmt.Fprintf(&b, "Size `%.4f` stop distance `%.4f` MPU `%.4f` expected risk `%.4f`.\n\n",
			a.PreSignal.PositionSize, a.PreSignal.StopDistance, a.PreSignal.MoneyPerPriceUnit, a.PreSignal.ExpectedAccountRisk)
		fmt.Fprintf(&b, "## What did broker execute?\n\n")
		fmt.Fprintf(&b, "See lifecycle events. Signal `%s`.\n\n", a.PreSignal.SignalID)
		fmt.Fprintf(&b, "## Were monetary semantics correct?\n\n")
		fmt.Fprintf(&b, "MPU at signal `%.4f`. Expected account-currency risk `%.4f`.\n\n",
			a.PreSignal.MoneyPerPriceUnit, a.PreSignal.ExpectedAccountRisk)
		fmt.Fprintf(&b, "## Were SL/TP correct?\n\n")
		fmt.Fprintf(&b, "Expected SL `%.4f` TP `%.4f`.\n\n", a.PreSignal.StopLoss, a.PreSignal.TakeProfit)
	}
	fmt.Fprintf(&b, "## Did local and broker state agree?\n\n")
	if a.Final != nil {
		fmt.Fprintf(&b, "Counts agree `%v` local `%d` broker `%d`.\n\n", a.Final.CountsAgree, a.Final.LocalPositions, a.Final.BrokerPositions)
		fmt.Fprintf(&b, "## How did trade exit?\n\n")
		fmt.Fprintf(&b, "Reason `%s` (not inferred from PnL). Trade outcome `%s`. Operational outcome `%s`.\n\n",
			a.Final.ExitReason, a.Final.TradeOutcome, a.Final.OperationalOutcome)
	} else {
		b.WriteString("Final reconciliation not yet written.\n\n")
	}
	b.WriteString("## What did global intelligence see?\n\n")
	b.WriteString("Intelligence files are `" + LabelObservational + "`. Not execution input.\n")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "REPORT.md"), []byte(b.String()), 0644)
}

func MustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
