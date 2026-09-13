package md

import (
	"context"
	"fmt"
)

const StatusNotConnected = "NOT_CONNECTED"

var (
	CME_NQ            = Placeholder{ID: "CME_NQ", Venue: "cme", Symbol: "NQ"}
	CME_GC            = Placeholder{ID: "CME_GC", Venue: "cme", Symbol: "GC"}
	NasdaqTotalView   = Placeholder{ID: "NASDAQ_TOTALVIEW", Venue: "nasdaq", Symbol: "TOTALVIEW"}
	Options           = Placeholder{ID: "OPTIONS", Venue: "opra", Symbol: "OPTIONS"}
)

type Placeholder struct {
	ID, Venue, Symbol string
}

func (p Placeholder) Name() string { return p.ID }

func (p Placeholder) Capabilities() Caps { return 0 }

func (p Placeholder) Status() string { return StatusNotConnected }

func (p Placeholder) Run(ctx context.Context, out *Bus) error {
	_ = ctx
	_ = out
	return fmt.Errorf("%s = %s (no credentials / no fake book)", p.ID, StatusNotConnected)
}

func PaidFeedStatuses() map[string]string {
	return map[string]string{
		"CME_NQ":            CME_NQ.Status(),
		"CME_GC":            CME_GC.Status(),
		"NASDAQ_TOTALVIEW":  NasdaqTotalView.Status(),
		"OPTIONS":           Options.Status(),
	}
}
