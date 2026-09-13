package globalsources

import "time"

type ReleaseInfo struct {
	Source       string    `json:"source"`
	Cadence      string    `json:"cadence"`
	LastActual   time.Time `json:"last_actual"`
	NextExpected time.Time `json:"next_expected"`
	Note         string    `json:"note"`
}

func NextWeekday(from time.Time, weekday time.Weekday) time.Time {
	t := from.UTC()
	for t.Weekday() != weekday {
		t = t.AddDate(0, 0, 1)
	}
	return t
}

func Calendar(now time.Time, last map[string]time.Time) []ReleaseInfo {
	now = now.UTC()
	out := []ReleaseInfo{
		{Source: "FED_H41", Cadence: "WEEKLY", Note: "H.4.1 typically Thursday"},
		{Source: "CFTC", Cadence: "WEEKLY", Note: "Disagg COT Friday"},
		{Source: "ICI", Cadence: "WEEKLY", Note: "weekly fund/ETF trends"},
		{Source: "TIC", Cadence: "MONTHLY", Note: "Treasury TIC ~6-8 weeks lag"},
		{Source: "JPX", Cadence: "WEEKLY", Note: "investor-type Thursday/Friday"},
	}
	for i := range out {
		if t, ok := last[out[i].Source]; ok {
			out[i].LastActual = t
		}
		switch out[i].Source {
		case "TIC":
			next := time.Date(now.Year(), now.Month(), 16, 0, 0, 0, 0, time.UTC)
			if !next.After(now) {
				next = next.AddDate(0, 1, 0)
			}
			out[i].NextExpected = next
		default:
			out[i].NextExpected = NextWeekday(now.AddDate(0, 0, 1), time.Thursday)
		}
	}
	return out
}
