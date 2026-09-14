package chronosplit

import "time"

type Split struct {
	DiscoveryEnd time.Time
	ValidEnd     time.Time
	HoldoutEnd   time.Time
	First        time.Time
	Last         time.Time
}

func Of(first, last time.Time) Split {
	if last.Before(first) {
		first, last = last, first
	}
	span := last.Sub(first)
	return Split{
		First: first, Last: last,
		DiscoveryEnd: first.Add(time.Duration(float64(span) * 0.60)),
		ValidEnd:     first.Add(time.Duration(float64(span) * 0.80)),
		HoldoutEnd:   last,
	}
}

func Bucket(t time.Time, s Split) string {
	if t.Before(s.First) || t.After(s.Last) {
		return ""
	}
	if !t.After(s.DiscoveryEnd) {
		return "DISCOVERY"
	}
	if !t.After(s.ValidEnd) {
		return "VALIDATION"
	}
	return "HOLDOUT"
}
