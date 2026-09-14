package money

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEvidenceCompleteRequiresRuntimeFields(t *testing.T) {
	s := MonetaryInstrumentSpec{ValidationStatus: RuntimeValidated, MoneyPerPriceUnit: 1, MoneyPerPriceUnitCurrency: "USD", CalibrationSamples: 80, Estimator: EstimatorTheilSen, PriceRange: 5.6, UPLRange: 0.005, MetadataStatus: MetadataRuntimeAgree, PositionClosed: true}
	if !EvidenceComplete(s) {
		t.Fatal("expected complete")
	}
	s.ValidationStatus = BrokerMetadataOnly
	if EvidenceComplete(s) {
		t.Fatal("metadata")
	}
}

func TestLoadCalibrationJournalAndComplete(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "US100.jsonl")
	line := `{"account_currency":"USD","ask":100.2,"bid":100,"closeablePrice":100,"dealId":"D1","direction":"BUY","size":0.001,"upl":-0.001,"instrument":"US100","event_time":"2026-09-14T01:16:47Z","receive_time":"2026-09-14T01:16:48Z"}` + "\n"
	if err := os.WriteFile(path, []byte(line), 0644); err != nil {
		t.Fatal(err)
	}
	samples, err := LoadCalibrationJournal(path)
	if err != nil || len(samples) != 1 || samples[0].DealID != "D1" {
		t.Fatalf("%+v %v", samples, err)
	}
	spec := CompleteEvidence(MonetaryInstrumentSpec{Epic: "US100"}, CalibrationResult{
		OK: true, MoneyPerPriceUnit: 1, Currency: "USD", Samples: 80, Estimator: EstimatorTheilSen,
		PriceRange: 5.6, UPLRange: 0.005, MetadataStatus: MetadataRuntimeAgree, EvidenceDealID: "D1",
	}, 0)
	if !EvidenceComplete(spec) || spec.BrokerPositionsAfter != 0 || !spec.PositionClosed {
		t.Fatalf("%+v", spec)
	}
}
