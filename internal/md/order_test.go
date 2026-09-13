package md

import (
	"testing"
	"time"
)

func TestEventOrderingBySeqThenTime(t *testing.T) {
	t0 := time.Unix(100, 0).UTC()
	events := []Event{
		{Kind: KindBookDelta, Seq: 3, EventTime: t0.Add(2 * time.Second)},
		{Kind: KindBookDelta, Seq: 1, EventTime: t0},
		{Kind: KindBookDelta, Seq: 2, EventTime: t0.Add(time.Second)},
	}
	ordered := OrderEvents(events)
	if ordered[0].Seq != 1 || ordered[1].Seq != 2 || ordered[2].Seq != 3 {
		t.Fatalf("%+v", ordered)
	}
}
