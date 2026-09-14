package stratrade

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type RiskRecord struct {
	Policy             string  `json:"policy"`
	Equity             float64 `json:"equity"`
	PerPositionCap     float64 `json:"per_position_cap"`
	AggregateCap       float64 `json:"aggregate_cap"`
	PlannedRisk        float64 `json:"planned_risk"`
	ActualModeledRisk  float64 `json:"actual_modeled_risk"`
	Difference         float64 `json:"difference"`
	StopDistance       float64 `json:"stop_distance"`
	Size               float64 `json:"size"`
	MoneyPerPriceUnit  float64 `json:"money_per_price_unit"`
	PilotStillHolds    bool    `json:"pilot_still_holds"`
}

func WriteRisk(root, id string, rec RiskRecord) error {
	if root == "" {
		return nil
	}
	dir := root
	if id != "" {
		dir = filepath.Join(root, id)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "risk.json"), b, 0644)
}

func WritePilotReport(root, id string, body string) error {
	if root == "" {
		return nil
	}
	dir := root
	if id != "" {
		dir = filepath.Join(root, id)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "REPORT.md"), []byte(body), 0644)
}

func FormatPilotReport(origin, market, sid string, planned, actual, slip float64, opsOut, tradeOut, exit string) string {
	return fmt.Sprintf("# %s %s\n\n- origin: %s\n- signal: %s\n- planned risk: %.4f\n- actual modeled risk: %.4f\n- slippage: %.4f\n- exit: %s\n- trade outcome: %s\n- operational outcome: %s\n\nObservational Attention/Coverage/Salience are not execution inputs.\n",
		market, sid, origin, sid, planned, actual, slip, exit, tradeOut, opsOut)
}
