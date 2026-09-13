package secctx

import (
	"testing"
	"time"
)

func TestNoLookaheadFiling(t *testing.T) {
	filed := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	f := []Filing{{Accession: "A", Form: "10-Q", FiledAt: filed, ReportPeriod: filed.AddDate(0, 0, -40), AvailableAt: filed.Add(24 * time.Hour), Source: "test"}}
	if LatestAvailable(f, filed) != nil {
		t.Fatal("not yet available")
	}
	if LatestAvailable(f, filed.Add(25*time.Hour)) == nil {
		t.Fatal("should be available")
	}
	p := ContextAt(f, filed.Add(48*time.Hour))
	if !p.UsableAt(filed.Add(48*time.Hour)) || p.UsableAt(filed) {
		t.Fatalf("%+v", p)
	}
}
