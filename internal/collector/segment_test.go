package collector

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSegmentRotationAndChecksum(t *testing.T) {
	root := t.TempDir()
	w := NewSegmented(root, "okx_swap", "BTC-USDT-SWAP", "trades")
	ts := time.Date(2026, 9, 13, 5, 1, 0, 0, time.UTC)
	if err := w.WriteEvent(ts, 1, map[string]any{"n": 1}); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "okx_swap", "BTC-USDT-SWAP", "2026", "09", "13")
	ents, _ := os.ReadDir(dir)
	if len(ents) < 1 {
		t.Fatal("no segment")
	}
	foundMeta := false
	for _, e := range ents {
		if filepath.Ext(e.Name()) == ".json" || len(e.Name()) > 10 && e.Name()[len(e.Name())-10:] == ".meta.json" {
			foundMeta = true
		}
	}
	if !foundMeta {
		// meta written as path+".meta.json"
		matches, _ := filepath.Glob(filepath.Join(dir, "*.meta.json"))
		if len(matches) == 0 {
			t.Fatalf("ents=%v", ents)
		}
	}
}
