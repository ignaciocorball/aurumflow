package strathist

import (
	"context"
	"time"

	"aurumflow/internal/market"
	"aurumflow/pkg/models"
)

type Book struct {
	Status                          string
	M5, M15, H1, H4                 []models.Candle
	M5Count, M15Count, H1Count, H4Count int
	M5Oldest, M15Oldest, H1Oldest, H4Oldest time.Time
	M5Latest, M15Latest, H1Latest, H4Latest time.Time
	WeekendGaps                     int
	UnexpectedGaps                  int
	ReadyAt                         time.Time
	FetchError                      string
}

type Fetcher interface {
	GetPrices(ctx context.Context, epic, resolution string, max int, from, to time.Time) ([]models.Candle, error)
}

func Seed(ctx context.Context, client Fetcher, epic string, req LegacyRequirements, asOf time.Time) Book {
	b := Book{Status: HistoryLoading}
	fetch := func(res string, lookback, period time.Duration, max int) []models.Candle {
		if client == nil {
			b.FetchError = "history provider unavailable"
			return nil
		}
		if max <= 0 {
			max = 400
		}
		raw, err := client.GetPrices(ctx, epic, res, max, asOf.Add(-lookback), asOf)
		if err != nil {
			b.FetchError = err.Error()
			return nil
		}
		return ClosedOnly(raw, asOf, period)
	}
	b.M5 = fetch(market.ResolutionMinute5, req.M5Lookback, req.M5Period, 500)
	b.M15 = fetch(market.ResolutionMinute15, req.M15Lookback, req.M15Period, 500)
	b.H1 = fetch(market.ResolutionHour, req.H1Lookback, req.H1Period, 400)
	b.H4 = fetch(market.ResolutionHour4, req.H4Lookback, req.H4Period, 400)
	b.Refresh(req)
	if b.Status == HistoryReady {
		b.ReadyAt = time.Now().UTC()
	}
	return b
}

func (b *Book) Refresh(req LegacyRequirements) {
	if b == nil {
		return
	}
	b.M5Count, b.M15Count, b.H1Count, b.H4Count = len(b.M5), len(b.M15), len(b.H1), len(b.H4)
	b.M5Oldest, b.M5Latest = OldestLatest(b.M5)
	b.M15Oldest, b.M15Latest = OldestLatest(b.M15)
	b.H1Oldest, b.H1Latest = OldestLatest(b.H1)
	b.H4Oldest, b.H4Latest = OldestLatest(b.H4)
	we, un := ClassifyGaps(b.M15, req.M15Period)
	b.WeekendGaps, b.UnexpectedGaps = we, un
	b.Status = Status(b.M5Count, b.M15Count, b.H1Count, b.H4Count, req, b.FetchError != "")
}

func (b *Book) MergeLive(liveM5, liveM15, liveH1, liveH4 []models.Candle, req LegacyRequirements, asOf time.Time) {
	if b == nil {
		return
	}
	b.M5 = mergeTF(b.M5, liveM5, asOf, req.M5Period)
	b.M15 = mergeTF(b.M15, liveM15, asOf, req.M15Period)
	b.H1 = mergeTF(b.H1, liveH1, asOf, req.H1Period)
	b.H4 = mergeTF(b.H4, liveH4, asOf, req.H4Period)
	b.Refresh(req)
	if b.Status == HistoryReady && b.ReadyAt.IsZero() {
		b.ReadyAt = time.Now().UTC()
	}
}

func mergeTF(seed, live []models.Candle, asOf time.Time, period time.Duration) []models.Candle {
	closedSeed := ClosedOnly(seed, asOf, period)
	var liveOK []models.Candle
	for _, c := range live {
		if !ValidOHLC(c) || c.Time.After(asOf) {
			continue
		}
		liveOK = append(liveOK, c)
	}
	return Merge(closedSeed, liveOK)
}
