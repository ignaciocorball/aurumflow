package market

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"aurumflow/pkg/models"
)

// PricesResponse is the response from GET /api/v1/prices/{epic}.
type PricesResponse struct {
	Prices         []PricePoint `json:"prices"`
	InstrumentType string       `json:"instrumentType"`
}

// PricePoint is a single OHLC point from the API (bid/ask separate).
type PricePoint struct {
	SnapshotTimeUTC string     `json:"snapshotTimeUTC"`
	SnapshotTime    string     `json:"snapshotTime"`
	OpenPrice       PriceLevel `json:"openPrice"`
	ClosePrice      PriceLevel `json:"closePrice"`
	HighPrice       PriceLevel `json:"highPrice"`
	LowPrice        PriceLevel `json:"lowPrice"`
	LastTradedVolume float64   `json:"lastTradedVolume"`
}

// PriceLevel holds bid/ask.
type PriceLevel struct {
	Bid float64 `json:"bid"`
	Ask float64 `json:"ask"`
}

// Resolution is the Capital.com resolution query value.
const (
	ResolutionMinute   = "MINUTE"
	ResolutionMinute5  = "MINUTE_5"
	ResolutionMinute15 = "MINUTE_15"
	ResolutionMinute30 = "MINUTE_30"
	ResolutionHour     = "HOUR"
	ResolutionHour4    = "HOUR_4"
	ResolutionDay      = "DAY"
	ResolutionWeek     = "WEEK"
)

// GetPrices fetches OHLC for the given epic and resolution, mapped to Candles (using bid for O/H/L/C).
func (c *Client) GetPrices(ctx context.Context, epic, resolution string, max int, from, to time.Time) ([]models.Candle, error) {
	path := fmt.Sprintf("/api/v1/prices/%s?resolution=%s&max=%d", epic, resolution, max)
	if !from.IsZero() {
		path += "&from=" + from.UTC().Format("2006-01-02T15:04:05")
	}
	if !to.IsZero() {
		path += "&to=" + to.UTC().Format("2006-01-02T15:04:05")
	}
	data, err := c.Do(ctx, "GET", path, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("get prices: %w", err)
	}
	var pr PricesResponse
	if err := json.Unmarshal(data, &pr); err != nil {
		return nil, fmt.Errorf("parse prices: %w", err)
	}
	candles := make([]models.Candle, 0, len(pr.Prices))
	for _, p := range pr.Prices {
		t, err := time.Parse("2006-01-02T15:04:05", p.SnapshotTimeUTC)
		if err != nil {
			t, _ = time.Parse(time.RFC3339, p.SnapshotTimeUTC)
		}
		candles = append(candles, models.Candle{
			Time:   t.UTC(),
			Open:   p.OpenPrice.Bid,
			High:   p.HighPrice.Bid,
			Low:    p.LowPrice.Bid,
			Close:  p.ClosePrice.Bid,
			Volume: p.LastTradedVolume,
		})
	}
	return candles, nil
}
