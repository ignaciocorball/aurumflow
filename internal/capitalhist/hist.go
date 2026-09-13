package capitalhist

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"aurumflow/internal/catalog"
	"aurumflow/internal/market"
	"aurumflow/pkg/models"
)

var CoreEpics = []string{"GOLD", "US100", "US500", "SILVER", "BTCUSD", "AAPL", "MSFT", "NVDA", "META", "AMZN"}

type HistoricalProvider struct {
	Client *market.Client
	Dir    string
}

func (p *HistoricalProvider) FetchWindow(ctx context.Context, epic, res string, from, to time.Time, max int) ([]models.Candle, error) {
	if max <= 0 {
		max = 500
	}
	var all []models.Candle
	cur := from
	for cur.Before(to) {
		end := cur.Add(7 * 24 * time.Hour)
		if end.After(to) {
			end = to
		}
		chunk, err := p.Client.GetPrices(ctx, epic, res, max, cur, end)
		if err != nil {
			return all, err
		}
		all = append(all, chunk...)
		if len(chunk) == 0 {
			cur = end
		} else {
			cur = chunk[len(chunk)-1].Time.Add(time.Second)
		}
		select {
		case <-ctx.Done():
			return all, ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
	}
	return DedupSort(all), nil
}

func DedupSort(in []models.Candle) []models.Candle {
	sort.Slice(in, func(i, j int) bool { return in[i].Time.Before(in[j].Time) })
	out := make([]models.Candle, 0, len(in))
	var last time.Time
	for _, c := range in {
		if c.Time.Equal(last) {
			continue
		}
		out = append(out, c)
		last = c.Time
	}
	return out
}

func CountGaps(cs []models.Candle, step time.Duration) int {
	if step <= 0 || len(cs) < 2 {
		return 0
	}
	n := 0
	for i := 1; i < len(cs); i++ {
		if cs[i].Time.Sub(cs[i-1].Time) > step+time.Minute {
			n++
		}
	}
	return n
}

func (p *HistoricalProvider) Save(epic, res string, cs []models.Candle) (string, error) {
	dir := filepath.Join(p.Dir, "capital", epic)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, res+".json")
	data, err := json.Marshal(cs)
	if err != nil {
		return "", err
	}
	return path, os.WriteFile(path, data, 0o644)
}

func Discover(ctx context.Context, c *market.Client, terms []string) (found, missing []string) {
	for _, t := range terms {
		ms, err := c.SearchMarkets(ctx, t)
		if err != nil || len(ms) == 0 {
			missing = append(missing, t)
			continue
		}
		ok := false
		for _, m := range ms {
			if m.Epic == t || m.InstrumentName != "" {
				found = append(found, m.Epic)
				ok = true
				break
			}
		}
		if !ok {
			missing = append(missing, t)
		}
	}
	return found, missing
}

func CatalogEntry(epic, res string, cs []models.Candle, path string) catalog.Dataset {
	from, to := "", ""
	if len(cs) > 0 {
		from = cs[0].Time.UTC().Format(time.RFC3339)
		to = cs[len(cs)-1].Time.UTC().Format(time.RFC3339)
	}
	return catalog.Dataset{
		Provider: "capital.com", Instrument: epic, Type: "ohlc:" + res,
		From: from, To: to, Rows: len(cs), Quality: "DIRECT",
		Gaps: CountGaps(cs, guessStep(res)), DownloadedAt: catalog.NowUTC(),
		SourceVersion: "capital-prices-v1", Path: path, Status: "ok",
	}
}

func guessStep(res string) time.Duration {
	switch res {
	case market.ResolutionMinute5:
		return 5 * time.Minute
	case market.ResolutionMinute15:
		return 15 * time.Minute
	case market.ResolutionHour:
		return time.Hour
	case market.ResolutionHour4:
		return 4 * time.Hour
	case market.ResolutionDay:
		return 24 * time.Hour
	default:
		return time.Hour
	}
}

func Fmt(n int) string { return fmt.Sprintf("%d", n) }
