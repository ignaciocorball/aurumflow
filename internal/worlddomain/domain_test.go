package worlddomain

import (
	"testing"
	"time"
)

func TestObservationNoLookahead(t *testing.T) {
	at := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	early := ContextObservation{Present: true, Value: 1, AvailableAt: at.Add(-time.Hour)}
	late := ContextObservation{Present: true, Value: 2, AvailableAt: at.Add(time.Hour)}
	missing := ContextObservation{Present: false, Value: 0, AvailableAt: at.Add(-time.Hour)}
	if !early.UsableAt(at) || late.UsableAt(at) || missing.UsableAt(at) {
		t.Fatal("lookahead or missing-as-present")
	}
}

func TestUnknownNotZero(t *testing.T) {
	if MissingLabel() == "0" || MissingLabel() == "" {
		t.Fatal(MissingLabel())
	}
	if HealthFromAge(time.Time{}, time.Now().UTC(), time.Hour, 24*time.Hour) != HealthUnknown {
		t.Fatal("zero time must be UNKNOWN")
	}
}

func TestStaleHealth(t *testing.T) {
	now := time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC)
	if HealthFromAge(now.Add(-2*time.Hour), now, time.Hour, 48*time.Hour) != HealthStale {
		t.Fatal("stale")
	}
	if HealthFromAge(now.Add(-30*time.Minute), now, time.Hour, 48*time.Hour) != HealthHealthy {
		t.Fatal("healthy")
	}
}
