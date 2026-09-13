package ctxsnap

import (
	"testing"
	"time"

	"aurumflow/internal/cftc"
	"aurumflow/internal/finra"
	"aurumflow/internal/macro"
	"aurumflow/internal/secctx"
)

func TestNoLookaheadSnapshot(t *testing.T) {
	filed := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	filings := []secctx.Filing{{
		Accession: "A", Form: "10-Q", FiledAt: filed,
		ReportPeriod: filed.AddDate(0, 0, -40), AvailableAt: filed.Add(24 * time.Hour), Source: "test",
	}}
	asOf := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	cot := []cftc.Row{{Market: cftc.GoldContract, AsOf: asOf, Available: cftc.AvailableAt(asOf), MMLong: 8, MMShort: 3, MMNet: 5}}
	weeks := []finra.Week{finra.ParseWeek("AAPL", "2026-08-07", 10, 2, true)}
	mac := &macro.Provider{Status: "TEST", Points: []macro.SeriesPoint{{
		ID: "DGS10", Period: filed, AvailableAt: filed.Add(6 * time.Hour), Value: 4.1, Freq: "D",
	}}}
	early := At(filed, filings, cot, weeks, "AAPL", "2026-08-07", mac, "DGS10")
	if early.SEC != nil {
		t.Fatal("sec lookahead")
	}
	late := At(filed.Add(48*time.Hour), filings, cot, weeks, "AAPL", "2026-08-07", mac, "DGS10")
	if late.SEC == nil || !late.CFTC.Present || late.FINRAMiss || !late.MacroOK {
		t.Fatalf("%+v", late)
	}
	missing := At(filed.Add(48*time.Hour), filings, cot, weeks, "MSFT", "2026-08-07", mac, "DGS10")
	if !missing.FINRAMiss {
		t.Fatal("missing must not become zero silently")
	}
}
