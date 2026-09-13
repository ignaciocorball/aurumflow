package macro

import (
	"os"
	"time"
)

type SeriesPoint struct {
	ID          string
	Period      time.Time
	AvailableAt time.Time
	Value       float64
	Freq        string
}

type Provider struct {
	Status string
	Points []SeriesPoint
}

func New() *Provider {
	p := &Provider{Status: "PENDING_FREE_KEY"}
	if os.Getenv("FRED_API_KEY") != "" {
		p.Status = "KEY_PRESENT_NOT_FETCHED"
	}
	return p
}

func (p *Provider) Latest(id string, at time.Time) (SeriesPoint, bool) {
	var best SeriesPoint
	ok := false
	for _, x := range p.Points {
		if x.ID != id || x.AvailableAt.After(at) {
			continue
		}
		if !ok || x.AvailableAt.After(best.AvailableAt) {
			best, ok = x, true
		}
	}
	return best, ok
}

var DefaultIDs = []string{"FEDFUNDS", "DGS2", "DGS10"}
