package md

import "time"

type DataQuality string
type FreshnessClass string

const (
	QualityDirect       DataQuality = "DIRECT"
	QualityProxy        DataQuality = "PROXY"
	QualityDelayed      DataQuality = "DELAYED"
	QualityExperimental DataQuality = "EXPERIMENTAL"

	FreshDirect     FreshnessClass = "DIRECT"
	FreshProxy      FreshnessClass = "PROXY"
	FreshDelayed    FreshnessClass = "DELAYED"
	FreshSlow       FreshnessClass = "SLOW_CONTEXT"
	FreshExperimental FreshnessClass = "EXPERIMENTAL"
)

// Provenance is attached to every normalized observation.
type Provenance struct {
	Provider       string
	Venue          string
	Instrument     string
	Quality        DataQuality
	FreshnessClass FreshnessClass
	EventTime      time.Time
	ReceivedAt     time.Time
	PublishedAt    *time.Time
	AvailableAt    time.Time
	IsProxy        bool
	IsDelayed      bool
}

func (p Provenance) UsableAt(t time.Time) bool {
	avail := p.AvailableAt
	if avail.IsZero() {
		if p.PublishedAt != nil {
			avail = *p.PublishedAt
		} else {
			avail = p.EventTime
		}
	}
	return !avail.After(t)
}

func PointerTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	u := t.UTC()
	return &u
}
