package worldstate

import (
	"testing"
	"time"

	"aurumflow/internal/worlddomain"
)

func TestWorldStateHashDeterministic(t *testing.T) {
	now := time.Date(2026, 9, 13, 16, 0, 0, 0, time.UTC)
	obs := []worlddomain.ContextObservation{{
		Source: "TIC", Metric: "NET", Value: 207.1, Present: true,
		Origin: worlddomain.OriginLive, AvailableAt: now.Add(-24 * time.Hour),
		RetrievedAt: now.Add(-time.Hour), PublishedAt: now.Add(-24 * time.Hour),
		ObservedAt: now.Add(-24 * time.Hour),
	}}
	a := At(now, Input{Observations: obs, Production: true})
	b := At(now, Input{Observations: obs, Production: true})
	if a.Hash == "" || a.Hash != b.Hash {
		t.Fatalf("hash %q vs %q", a.Hash, b.Hash)
	}
	obs2 := append([]worlddomain.ContextObservation{}, obs...)
	obs2[0].Value = 100
	c := At(now, Input{Observations: obs2, Production: true})
	if c.Hash == "" || c.Hash == a.Hash {
		t.Fatal("evidence change must change hash")
	}
}
