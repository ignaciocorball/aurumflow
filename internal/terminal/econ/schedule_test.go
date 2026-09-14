package econ

import (
	"testing"
	"time"
)

func TestEventCountdownAndWindow(t *testing.T) {
	now := time.Date(2026, 9, 14, 16, 18, 0, 0, time.UTC)
	e := Event{Name: "US CPI", At: now.Add(time.Hour + 42*time.Minute), Importance: "HIGH"}
	st, window, cd := StatusOf(e, now)
	if st != "UPCOMING" || window != "PRE_EVENT" {
		t.Fatalf("%s %s %s", st, window, cd)
	}
	if cd == "" {
		t.Fatal("countdown")
	}
	active := Event{At: now.Add(10 * time.Minute)}
	st, window, _ = StatusOf(active, now)
	if st != "ACTIVE_WINDOW" || window != "ACTIVE" {
		t.Fatalf("active %s %s", st, window)
	}
	rel := Event{At: now.Add(-time.Hour)}
	st, window, _ = StatusOf(rel, now)
	if st != "RELEASED" || window != "POST_EVENT" {
		t.Fatalf("released %s %s", st, window)
	}
}

func TestEIAWednesdaysHaveSource(t *testing.T) {
	now := time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC)
	ev := eiaWednesdays(now)
	if len(ev) == 0 {
		t.Fatal("none")
	}
	if ev[0].SourceURL == "" || ev[0].Source == "" {
		t.Fatal("provenance")
	}
}
