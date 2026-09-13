package worldstate

import (
	"testing"
	"time"

	"aurumflow/internal/cftc"
	"aurumflow/internal/globalsources"
	"aurumflow/internal/worlddomain"
	"aurumflow/internal/xasset"
)

func TestWorldStateDeterministicAndNoLookahead(t *testing.T) {
	now := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	obs := globalsources.LoadFixtureDir("../globalsources/testdata", now)
	if len(obs) == 0 {
		t.Fatal("fixtures")
	}
	cot := []cftc.Row{{
		Market: cftc.GoldContract, AsOf: now.AddDate(0, 0, -10), Available: now.AddDate(0, 0, -6),
		MMLong: 20, MMShort: 8, MMNet: 12,
	}}
	in := Input{Observations: obs, COT: cot, BTCMicro: true, Frames: []xasset.Frame{
		{Symbol: "GOLD", Close: []float64{1900, 1920, 1950, 1980}},
		{Symbol: "US500", Close: []float64{5000, 5010, 5005, 5020}},
	}, US500: []float64{5000, 5010, 5005, 5020}, MegaDirs: []int{1, 1, -1, 1, 1}}
	a := At(now, in)
	b := At(now, in)
	if a.Liquidity.Label != "HEURISTIC_V1" || a.Risk == "" {
		t.Fatalf("%+v", a)
	}
	if a.AsOf != b.AsOf || a.Liquidity.Class != b.Liquidity.Class {
		t.Fatal("nondeterministic")
	}
	early := At(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), in)
	if early.Markets["GOLD"].CapitalFlowContext == "INFLOW" {
		t.Fatal("lookahead gold flow")
	}
	if a.Regions[worlddomain.RegionUS].Region == "" {
		t.Fatal("region")
	}
	if a.AssetClasses[worlddomain.AssetPrecious].Class == "" {
		t.Fatal("asset")
	}
	if a.Confidence <= 0 {
		t.Fatal(a.Confidence)
	}
	stale := append([]worlddomain.ContextObservation{}, obs...)
	if o, ok := globalsources.Latest(stale, "ICI", "EQUITY", now); ok {
		o.Quality = worlddomain.HealthStale
		stale = append(stale, o)
	}
	low := At(now, Input{Observations: stale})
	if low.Confidence > a.Confidence {
		t.Fatal("stale should not raise confidence")
	}
	if a.Markets["GOLD"].Market == "" {
		t.Fatal("gold market")
	}
}

func TestUnknownNotZeroMarket(t *testing.T) {
	ws := At(time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC), Input{})
	if ws.USD.USD != "UNKNOWN" && ws.USD.USD != "" {
		t.Fatal(ws.USD)
	}
	if ws.Markets["BTC"].PriceTrend == "0" {
		t.Fatal("zero")
	}
}
