package research

import (
	"testing"
	"time"
)

func TestWalkForwardIsolation(t *testing.T) {
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(90 * 24 * time.Hour)
	s := ProportionalSplit(from, to)
	midValid := s.ValidFrom.Add(time.Hour)
	if InRange(midValid, s.TrainFrom, s.TrainTo) {
		t.Fatal("validation leaked into train")
	}
	if InRange(s.OOSFrom, s.TrainFrom, s.TrainTo) || InRange(s.OOSFrom, s.ValidFrom, s.ValidTo) {
		t.Fatal("oos leaked")
	}
	if !s.OOSTo.Equal(to) {
		t.Fatal("oos end")
	}
}
