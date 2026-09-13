package ops

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const GoldObservatoryFile = "journals/gold-observatory.json"

type GoldObservatory struct {
	MarketStatus  string  `json:"market_status"`
	Validation    string  `json:"validation"`
	QuotesOK      bool    `json:"quotes_ok"`
	Bid           float64 `json:"bid,omitempty"`
	Ask           float64 `json:"ask,omitempty"`
	Spread        float64 `json:"spread,omitempty"`
	DemoBalance   float64 `json:"demo_balance,omitempty"`
	DemoBalanceOK bool    `json:"demo_balance_ok,omitempty"`
	OpenPositions int     `json:"open_positions"`
	PositionsOK   bool    `json:"positions_ok"`
	Updated       string  `json:"updated"`
}

func WriteGoldObservatory(path string, g GoldObservatory) error {
	if path == "" {
		path = GoldObservatoryFile
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if g.Updated == "" {
		g.Updated = time.Now().UTC().Format(time.RFC3339)
	}
	if !g.QuotesOK {
		g.Bid, g.Ask, g.Spread = 0, 0, 0
	}
	raw, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func ReadGoldObservatory(path string) (GoldObservatory, bool) {
	if path == "" {
		path = GoldObservatoryFile
	}
	var g GoldObservatory
	raw, err := os.ReadFile(path)
	if err != nil {
		return g, false
	}
	if json.Unmarshal(raw, &g) != nil {
		return g, false
	}
	g.QuotesOK = GoldQuotesOK(g.MarketStatus, g.Bid, g.Ask)
	return g, g.MarketStatus != ""
}
