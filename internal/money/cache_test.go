package money

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSpecCacheRoundTripAndInvalidate(t *testing.T) {
	dir := t.TempDir()
	s := MonetaryInstrumentSpec{Epic: "GOLD", Currency: "USD", LotSize: 1, MinDealSize: 0.01, MaxDealSize: 10, SizeIncrement: 0.01, ValidationStatus: BrokerMetadataOnly}
	path := CachePath(dir, "gold")
	if err := SaveSpec(path, s); err != nil {
		t.Fatal(err)
	}
	got, err := LoadSpec(path)
	if err != nil || got.Epic != "GOLD" || got.EvidenceVersion != EvidenceVersion {
		t.Fatalf("%+v %v", got, err)
	}
	live := s
	live.LotSize = 10
	if !MetadataChanged(got, live) || InvalidateReason(got, live) == "" {
		t.Fatal("expected invalidate")
	}
	if _, err := os.Stat(filepath.Join(dir, "GOLD.json")); err != nil {
		t.Fatal(err)
	}
}
