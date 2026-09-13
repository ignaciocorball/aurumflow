package sessobs

import (
	"testing"
	"time"

	"aurumflow/internal/sessions"
)

func TestHandoffDescriptiveOnly(t *testing.T) {
	t0 := time.Date(2026, 9, 14, 13, 0, 0, 0, time.UTC)
	o, ok := Observe(sessions.Europe, sessions.US, t0, "DE40", "FLAT", "FLAT", "EUROPE")
	if !ok || o.AfterHandoff != "UNKNOWN" {
		t.Fatal(o, ok)
	}
	if _, ok := Observe(sessions.US, sessions.US, t0, "", "", "", ""); ok {
		t.Fatal("same phase")
	}
}
