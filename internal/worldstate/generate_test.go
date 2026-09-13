package worldstate

import (
	"context"
	"testing"
	"time"

	"aurumflow/internal/worlddomain"
)

func TestCachedOfficialNeverLoadsFixtureOrigins(t *testing.T) {
	in := LoadCachedOfficialInput(context.Background())
	if !in.Production {
		t.Fatal("official cache input must be production")
	}
	for _, o := range in.Observations {
		if o.Origin == worlddomain.OriginFixture {
			t.Fatalf("fixture origin in official cache %s %s", o.Source, o.Metric)
		}
	}
}

func TestOfficialFallbackIsCacheOrUnknown(t *testing.T) {
	in := LoadCachedOfficialInput(context.Background())
	ws := At(time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC), in)
	if ws.Valid == "FIXTURE_OR_TEST" || ws.Valid == "FIXTURE_CONTAMINATED" {
		t.Fatal(ws.Valid)
	}
	if !in.Production {
		t.Fatal("production")
	}
}
