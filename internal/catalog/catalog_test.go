package catalog

import (
	"path/filepath"
	"testing"
)

func TestCatalogRoundTrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "catalog.json")
	c, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	c.Upsert(Dataset{Provider: "binance_vision", Instrument: "BTCUSDT", Type: "aggTrades", From: "a", To: "b", Rows: 1})
	if err := c.Save(p); err != nil {
		t.Fatal(err)
	}
	c2, err := Load(p)
	if err != nil || len(c2.Datasets) != 1 {
		t.Fatal(err, c2)
	}
}
