package worldstate

import "time"

func DailyAsOf(start, end time.Time, step time.Duration) []time.Time {
	if step <= 0 {
		step = 24 * time.Hour
	}
	var out []time.Time
	for t := start.UTC(); !t.After(end.UTC()); t = t.Add(step) {
		out = append(out, t)
	}
	return out
}

func Replay(in Input, points []time.Time) []WorldState {
	out := make([]WorldState, 0, len(points))
	for _, t := range points {
		out = append(out, At(t, in))
	}
	return out
}
