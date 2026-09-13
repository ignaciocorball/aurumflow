package cftc

import (
	"testing"
	"time"
)

func TestGoldIdentityAndAvailability(t *testing.T) {
	if MatchGold("GOLDEN") || !MatchGold(GoldContract) {
		t.Fatal("identity")
	}
	asOf := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC) // Tuesday
	avail := AvailableAt(asOf)
	if avail.Before(asOf.Add(3 * 24 * time.Hour)) {
		t.Fatal(avail)
	}
	hist := []Row{{Market: GoldContract, AsOf: asOf, Available: avail, MMLong: 10, MMShort: 4, MMNet: 6}}
	if _, ok := Features(hist, asOf); ok {
		t.Fatal("lookahead")
	}
	got, ok := Features(hist, avail)
	if !ok || got.MMNet != 6 {
		t.Fatal(got, ok)
	}
	ctx := ContextAt(hist, avail)
	if !ctx.Present || ctx.Crowding <= 0 {
		t.Fatalf("%+v", ctx)
	}
}

func TestExactMarketIdentityNoFuzzy(t *testing.T) {
	if _, ok := CanonicalMarket("SILVERISH"); ok {
		t.Fatal("fuzzy")
	}
	if id, ok := CanonicalMarket(SilverContract); !ok || id != "SILVER" {
		t.Fatal(id, ok)
	}
	asOf := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)
	avail := AvailableAt(asOf)
	hist := []Row{{Market: SilverContract, AsOf: asOf, Available: avail, MMLong: 8, MMShort: 3, MMNet: 5}}
	if ContextFor(hist, "GOLD", avail).Present {
		t.Fatal("cross market")
	}
	got := ContextFor(hist, "SILVER", avail)
	if !got.Present || got.MMNet != 5 {
		t.Fatalf("%+v", got)
	}
}
