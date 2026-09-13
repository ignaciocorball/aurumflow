package research

import (
	"testing"
	"time"
)

func TestUntouchedExclusion(t *testing.T) {
	r := DefaultDatasetRegistry()
	if err := r.ForbidUntouched(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("must block pre-holdout untouched")
	}
	if err := r.ForbidUntouched(time.Date(2026, 3, 17, 0, 0, 0, 0, time.UTC), time.Date(2026, 6, 14, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
}
