package research

import "time"

type Split struct {
	TrainFrom, TrainTo time.Time
	ValidFrom, ValidTo time.Time
	OOSFrom, OOSTo     time.Time
}

func ProportionalSplit(from, to time.Time) Split {
	from, to = from.UTC(), to.UTC()
	span := to.Sub(from)
	if span <= 0 {
		return Split{TrainFrom: from, TrainTo: to, ValidFrom: to, ValidTo: to, OOSFrom: to, OOSTo: to}
	}
	trainEnd := from.Add(time.Duration(float64(span) * 60 / 90))
	validEnd := from.Add(time.Duration(float64(span) * 75 / 90))
	return Split{
		TrainFrom: from, TrainTo: trainEnd,
		ValidFrom: trainEnd, ValidTo: validEnd,
		OOSFrom: validEnd, OOSTo: to,
	}
}

func InRange(t, from, to time.Time) bool {
	return !t.Before(from) && t.Before(to)
}
