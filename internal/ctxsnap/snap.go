package ctxsnap

import (
	"time"

	"aurumflow/internal/cftc"
	"aurumflow/internal/finra"
	"aurumflow/internal/macro"
	"aurumflow/internal/md"
	"aurumflow/internal/secctx"
)

// Instant is the latest slow context usable at research time t (no lookahead).
type Instant struct {
	At         time.Time
	SEC        *secctx.Filing
	SECProv    md.Provenance
	CFTC       cftc.Context
	FINRAATS   float64
	FINRAMiss  bool
	Macro      macro.SeriesPoint
	MacroOK    bool
}

func At(t time.Time, filings []secctx.Filing, cot []cftc.Row, weeks []finra.Week, symbol, weekEnding string, mac *macro.Provider, seriesID string) Instant {
	out := Instant{At: t.UTC()}
	out.SEC = secctx.LatestAvailable(filings, t)
	out.SECProv = secctx.ContextAt(filings, t)
	out.CFTC = cftc.ContextAt(cot, t)
	if weekEnding != "" {
		_, _, _, _, _, _, miss := finra.Feature(weeks, symbol, weekEnding)
		out.FINRAMiss = miss
		if !miss {
			ats, _, _, _, _, _, _ := finra.Feature(weeks, symbol, weekEnding)
			out.FINRAATS = ats
		}
	} else {
		out.FINRAMiss = true
	}
	if mac != nil && seriesID != "" {
		out.Macro, out.MacroOK = mac.Latest(seriesID, t)
	}
	return out
}
