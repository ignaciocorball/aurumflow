package worldstate

import (
	"context"
	"time"

	"aurumflow/internal/cftc"
	"aurumflow/internal/globalsources"
	"aurumflow/internal/xasset"
)

func LoadResearchInput(fixtureDir, cacheDir string, now time.Time) Input {
	obs := globalsources.LoadFixtureDir(fixtureDir, now)
	if cacheDir != "" {
		obs = append(obs, globalsources.LoadFixtureDir(cacheDir, now)...)
	}
	return Input{
		Observations: obs,
		COT: []cftc.Row{{
			Market: cftc.GoldContract, AsOf: now.AddDate(0, 0, -10), Available: now.AddDate(0, 0, -6),
			MMLong: 20, MMShort: 8, MMNet: 12, PMNet: 4, SDNet: -2, ORNet: 1,
		}},
		BTCMicro: true,
		Frames: []xasset.Frame{
			{Symbol: "GOLD", Close: []float64{1900, 1920, 1950, 1980}},
			{Symbol: "US100", Close: []float64{18000, 18100, 18050, 18200}},
			{Symbol: "US500", Close: []float64{5000, 5010, 5005, 5020}},
			{Symbol: "BTC", Close: []float64{70000, 71000, 70500, 72000}},
		},
		US500:    []float64{5000, 5010, 5005, 5020},
		MegaDirs: []int{1, 1, -1, 1, 1},
	}
}

func LoadOfficialInput(ctx context.Context, now time.Time) (Input, []globalsources.FetchResult) {
	_ = now
	obs, res := globalsources.FetchOfficial(ctx)
	if len(obs) == 0 {
		// Allowed fallback is last CACHE_OFFICIAL only. Never testdata fixtures.
		obs = globalsources.LoadOfficialCache()
	}
	return Input{Observations: obs, Production: true}, res
}

func LoadCachedOfficialInput(ctx context.Context) Input {
	_ = ctx
	return Input{Observations: globalsources.LoadOfficialCache(), Production: true}
}
