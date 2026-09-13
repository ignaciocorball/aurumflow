package md

import (
	"testing"
	"time"
)

func TestNoLookaheadAvailableAt(t *testing.T) {
	event := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)   // COT Tuesday
	pub := time.Date(2026, 9, 11, 21, 30, 0, 0, time.UTC) // Friday publish
	p := Provenance{EventTime: event, PublishedAt: PointerTime(pub), AvailableAt: pub, FreshnessClass: FreshSlow}
	if p.UsableAt(event) {
		t.Fatal("cannot use COT on event Tuesday")
	}
	if p.UsableAt(pub.Add(-time.Hour)) {
		t.Fatal("cannot use before publish")
	}
	if !p.UsableAt(pub) {
		t.Fatal("usable at available_at")
	}
}
