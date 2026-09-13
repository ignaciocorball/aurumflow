package globalsources

import (
	"testing"
	"time"

	"aurumflow/internal/worlddomain"
)

func TestOfficialCoverageNoFabrication(t *testing.T) {
	now := time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC)
	obs := []worlddomain.ContextObservation{
		{Source: "FED_H41", Metric: "FED_ASSETS", Present: true, ObservedAt: now.AddDate(0, 0, -7)},
		{Source: "FED_H41", Metric: "FED_ASSETS", Present: true, ObservedAt: now.AddDate(0, 0, -14)},
	}
	rows := OfficialCoverage(obs, now)
	var fed CoverageRow
	for _, r := range rows {
		if r.Source == "FED_H41" {
			fed = r
		}
	}
	if fed.Have != 2 || fed.Want != 52 {
		t.Fatalf("%+v", fed)
	}
}

func TestCalendarExists(t *testing.T) {
	cal := Calendar(time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC), nil)
	if len(cal) == 0 {
		t.Fatal("calendar")
	}
}
