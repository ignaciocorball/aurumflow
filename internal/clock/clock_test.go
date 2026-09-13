package clock

import (
	"testing"
	"time"
)

func TestReplayClockDeterministic(t *testing.T) {
	c := &ReplayClock{}
	ts := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	c.Set(ts)
	if !c.Now().Equal(ts) {
		t.Fatal(c.Now())
	}
	c.Set(ts.Add(time.Minute))
	if c.Now().Equal(ts) {
		t.Fatal("must advance only when set")
	}
}
