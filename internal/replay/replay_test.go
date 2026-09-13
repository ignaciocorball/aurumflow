package replay

import (
	"testing"
	"time"

	"aurumflow/internal/clock"
	"aurumflow/internal/md"
)

func TestReplayDeterministicOrderAndClock(t *testing.T) {
	t0 := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	in := []md.Event{
		{Kind: md.KindTrade, EventTime: t0.Add(2 * time.Second), Seq: 2, Provider: "b"},
		{Kind: md.KindTrade, EventTime: t0, Seq: 9, Provider: "a"},
		{Kind: md.KindTrade, EventTime: t0.Add(2 * time.Second), Seq: 1, Provider: "a"},
	}
	eng := New()
	var times []time.Time
	var seqs []int64
	eng.Run(in, func(ev md.Event, clk *clock.ReplayClock) {
		if !clk.Now().Equal(ev.EventTime) {
			t.Fatalf("clock %v event %v", clk.Now(), ev.EventTime)
		}
		times = append(times, ev.EventTime)
		seqs = append(seqs, ev.Seq)
	})
	if !times[0].Equal(t0) || seqs[1] != 1 || seqs[2] != 2 {
		t.Fatalf("%v %v", times, seqs)
	}
	var seqs2 []int64
	eng.Run(in, func(ev md.Event, _ *clock.ReplayClock) { seqs2 = append(seqs2, ev.Seq) })
	if len(seqs2) != 3 || seqs2[0] != seqs[0] || seqs2[1] != seqs[1] {
		t.Fatal("nondeterministic")
	}
}
