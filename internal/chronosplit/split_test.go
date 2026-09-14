package chronosplit

import (
	"testing"
	"time"
)

func TestChronologicalNoShuffle(t *testing.T) {
	a := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	b := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)
	s := Of(a, b)
	if Bucket(a.Add(10*24*time.Hour), s) != "DISCOVERY" {
		t.Fatal("early")
	}
	if Bucket(s.DiscoveryEnd.Add(time.Hour), s) != "VALIDATION" {
		t.Fatal("mid")
	}
	if Bucket(s.ValidEnd.Add(time.Hour), s) != "HOLDOUT" {
		t.Fatal("late")
	}
	if s.DiscoveryEnd.After(s.ValidEnd) || s.ValidEnd.After(s.HoldoutEnd) {
		t.Fatal("order")
	}
}

func TestNoFutureLeak(t *testing.T) {
	a := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s := Of(a, a.Add(100*24*time.Hour))
	if Bucket(a.Add(200*24*time.Hour), s) != "" {
		t.Fatal("future bar must not join a split")
	}
}
