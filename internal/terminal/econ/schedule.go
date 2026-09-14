package econ

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type Event struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Region       string   `json:"region"`
	At           time.Time `json:"at"`
	Importance   string   `json:"importance"`
	Source       string   `json:"source"`
	SourceURL    string   `json:"source_url"`
	AssetClasses []string `json:"asset_classes"`
	Markets      []string `json:"markets"`
	Actual       string   `json:"actual,omitempty"`
	Forecast     string   `json:"forecast,omitempty"`
	Previous     string   `json:"previous,omitempty"`
}

type fileEvent struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Region       string   `json:"region"`
	At           string   `json:"at"`
	Importance   string   `json:"importance"`
	Source       string   `json:"source"`
	SourceURL    string   `json:"source_url"`
	AssetClasses []string `json:"asset_classes"`
	Markets      []string `json:"markets"`
}

func LoadSchedule(path string, now time.Time) []Event {
	var out []Event
	if b, err := os.ReadFile(path); err == nil {
		var rows []fileEvent
		if json.Unmarshal(b, &rows) == nil {
			for _, r := range rows {
				t, err := time.Parse(time.RFC3339, r.At)
				if err != nil {
					continue
				}
				out = append(out, Event{
					ID: r.ID, Name: r.Name, Region: r.Region, At: t.UTC(),
					Importance: r.Importance, Source: r.Source, SourceURL: r.SourceURL,
					AssetClasses: r.AssetClasses, Markets: r.Markets,
				})
			}
		}
	}
	out = append(out, eiaWednesdays(now)...)
	out = append(out, blsFirstFridays(now)...)
	return Dedup(out)
}

func Dedup(in []Event) []Event {
	seen := map[string]bool{}
	var out []Event
	for _, e := range in {
		key := e.ID
		if key == "" {
			key = e.Name + "|" + e.At.UTC().Format(time.RFC3339)
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, e)
	}
	return out
}

func StatusOf(e Event, now time.Time) (status, window, countdown string) {
	now = now.UTC()
	at := e.At.UTC()
	pre := at.Add(-15 * time.Minute)
	post := at.Add(30 * time.Minute)
	switch {
	case now.Before(pre):
		status, window = "UPCOMING", "PRE_EVENT"
	case now.Before(post):
		status, window = "ACTIVE_WINDOW", "ACTIVE"
	default:
		status, window = "RELEASED", "POST_EVENT"
	}
	d := at.Sub(now)
	if d < 0 {
		d = -d
		countdown = "released " + fmtDuration(d) + " ago"
	} else {
		countdown = fmtDuration(d)
	}
	return
}

func fmtDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h >= 48 {
		return fmt.Sprintf("%dd %dh", h/24, h%24)
	}
	if h >= 1 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if m >= 1 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}

func eiaWednesdays(now time.Time) []Event {
	var out []Event
	t := time.Date(now.Year(), now.Month(), now.Day(), 14, 30, 0, 0, time.UTC)
	for i := 0; i < 21; i++ {
		d := t.AddDate(0, 0, i)
		if d.Weekday() == time.Wednesday && d.After(now.Add(-7*24*time.Hour)) {
			out = append(out, Event{
				ID: "eia-wpsr-" + d.Format("20060102"),
				Name: "EIA Weekly Petroleum Status Report",
				Region: "UNITED_STATES", At: d,
				Importance: "HIGH",
				Source: "U.S. Energy Information Administration",
				SourceURL: "https://www.eia.gov/petroleum/supply/weekly/",
				AssetClasses: []string{"ENERGY"},
				Markets: []string{"OIL", "US500"},
			})
		}
	}
	return out
}

func blsFirstFridays(now time.Time) []Event {
	var out []Event
	for i := 0; i < 4; i++ {
		month := time.Date(now.Year(), now.Month()+time.Month(i), 1, 12, 30, 0, 0, time.UTC)
		d := month
		for d.Weekday() != time.Friday {
			d = d.AddDate(0, 0, 1)
		}
		out = append(out, Event{
			ID: "bls-empsit-" + d.Format("20060102"),
			Name: "Employment Situation",
			Region: "UNITED_STATES", At: d,
			Importance: "HIGH",
			Source: "U.S. Bureau of Labor Statistics",
			SourceURL: "https://www.bls.gov/schedule/news_release/empsit.htm",
			AssetClasses: []string{"EQUITIES", "PRECIOUS_METALS"},
			Markets: []string{"US100", "US500", "US30", "GOLD"},
		})
	}
	return out
}

func FilterWindow(events []Event, now time.Time, past, future time.Duration) []Event {
	lo := now.Add(-past)
	hi := now.Add(future)
	var out []Event
	for _, e := range events {
		if e.At.Before(lo) || e.At.After(hi) {
			continue
		}
		out = append(out, e)
	}
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].At.Before(out[i].At) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

func RegionLabel(r string) string {
	switch strings.ToUpper(r) {
	case "UNITED_STATES":
		return "United States"
	case "EUROPE":
		return "Europe"
	case "JAPAN":
		return "Japan"
	case "UNITED_KINGDOM":
		return "United Kingdom"
	default:
		return r
	}
}
