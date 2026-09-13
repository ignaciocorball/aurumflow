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
