package worldstate

import (
	"testing"
	"time"

	"aurumflow/internal/worlddomain"
)

func TestProductionRejectsFixture(t *testing.T) {
	now := time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC)
	fix := worlddomain.ContextObservation{
		Source: "TIC", Metric: "NET_LT_SECURITIES", Value: 12, Present: true,
		ObservedAt: now.AddDate(0, -2, 0), PublishedAt: now.AddDate(0, -2, 0), AvailableAt: now.AddDate(0, -1, 0),
		Origin: worlddomain.OriginFixture,
	}
	live := fix
	live.Origin = worlddomain.OriginLive
	live.Value = 3
	ws := At(now, Input{Production: true, Observations: []worlddomain.ContextObservation{fix, live}})
	if ws.RejectedFix == 0 || ws.Valid != "CURRENT_WORLD_STATE_VALID" {
		t.Fatalf("rejected=%d valid=%s", ws.RejectedFix, ws.Valid)
	}
}

func TestFuturePublicationRejected(t *testing.T) {
	now := time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC)
	o := worlddomain.ContextObservation{
		Source: "TIC", Metric: "NET_LT_SECURITIES", Value: 99, Present: true,
		ObservedAt: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		PublishedAt: now.Add(24 * time.Hour),
		AvailableAt: now.Add(24 * time.Hour),
		Origin: worlddomain.OriginLive,
	}
	if o.UsableAt(now) {
		t.Fatal("future publication usable")
	}
	ws := At(now, Input{Production: true, Observations: []worlddomain.ContextObservation{o}})
	if ws.USD.Present {
		t.Fatal("future TIC leaked")
	}
}

func TestEmptyOriginRejectedInProduction(t *testing.T) {
	now := time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC)
	o := worlddomain.ContextObservation{
		Source: "TIC", Metric: "NET_LT_SECURITIES", Value: 1, Present: true,
		ObservedAt: now.AddDate(0, -2, 0), AvailableAt: now.AddDate(0, -1, 0),
	}
	ws := At(now, Input{Production: true, Observations: []worlddomain.ContextObservation{o}})
	if ws.RejectedFix == 0 {
		t.Fatal("empty origin not rejected")
	}
	if ws.USD.Present {
		t.Fatal("empty-origin observation leaked")
	}
}

func TestCurrentWorldStateSourceAudit(t *testing.T) {
	now := time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC)
	o := worlddomain.ContextObservation{
		Source: "FED_H41", Metric: "FED_ASSETS", Value: 6600, Present: true,
		ObservedAt: now.AddDate(0, 0, -7), PublishedAt: now.AddDate(0, 0, -5), AvailableAt: now.AddDate(0, 0, -5),
		Origin: worlddomain.OriginLive,
	}
	ws := At(now, Input{Production: true, Observations: []worlddomain.ContextObservation{o}})
	if ws.Valid != "CURRENT_WORLD_STATE_VALID" {
		t.Fatal(ws.Valid)
	}
	if len(ws.Origins) != 1 || ws.Origins[0] != string(worlddomain.OriginLive) {
		t.Fatal(ws.Origins)
	}
}

func TestLiveAndCacheAccepted(t *testing.T) {
	now := time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC)
	o := worlddomain.ContextObservation{
		Source: "FED_H41", Metric: "FED_ASSETS", Value: 6600, Present: true,
		ObservedAt: now.AddDate(0, 0, -7), PublishedAt: now.AddDate(0, 0, -5), AvailableAt: now.AddDate(0, 0, -5),
		Origin: worlddomain.OriginCache,
	}
	ws := At(now, Input{Production: true, Observations: []worlddomain.ContextObservation{o}})
	if ws.RejectedFix != 0 || ws.Valid != "CURRENT_WORLD_STATE_VALID" {
		t.Fatalf("%+v", ws)
	}
}
