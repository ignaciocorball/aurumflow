package market

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// MarketsResponse is the response from GET /api/v1/markets?searchTerm=...
type MarketsResponse struct {
	Markets []MarketInfo `json:"markets"`
}

// MarketInfo holds epic and dealing rules.
type MarketInfo struct {
	Epic            string `json:"epic"`
	InstrumentName  string `json:"instrumentName"`
	InstrumentType  string `json:"instrumentType"`
	MarketStatus    string `json:"marketStatus"`
	MinDealSize     *SizeRule `json:"minDealSize,omitempty"`
	MaxDealSize     *SizeRule `json:"maxDealSize,omitempty"`
	MinSizeIncrement *SizeRule `json:"minSizeIncrement,omitempty"`
}

// SizeRule is used in dealing rules.
type SizeRule struct {
	Unit  string  `json:"unit"`
	Value float64 `json:"value"`
}

// MarketDetailsResponse is the response from GET /api/v1/markets/{epic}.
type MarketDetailsResponse struct {
	Instrument    MarketInstrument `json:"instrument"`
	DealingRules  DealingRules     `json:"dealingRules"`
	Snapshot      MarketSnapshot   `json:"snapshot"`
}

// MarketInstrument holds epic and metadata.
type MarketInstrument struct {
	Epic          string  `json:"epic"`
	Name          string  `json:"name"`
	LotSize       float64 `json:"lotSize"`
	Type          string  `json:"type"`
	Currency      string  `json:"currency,omitempty"`
	ScalingFactor float64 `json:"scalingFactor,omitempty"`
	ValueOfOnePip json.RawMessage `json:"valueOfOnePip,omitempty"`
	ContractSize  float64 `json:"contractSize,omitempty"`
	MarginFactor  float64 `json:"marginFactor,omitempty"`
	PipPosition   int     `json:"pipPosition,omitempty"`
	TickSize      float64 `json:"tickSize,omitempty"`
}

// DealingRules holds min/max deal size etc.
type DealingRules struct {
	MinDealSize              SizeRule `json:"minDealSize"`
	MaxDealSize              SizeRule `json:"maxDealSize"`
	MinSizeIncrement         SizeRule `json:"minSizeIncrement"`
	MinStopOrProfitDistance  SizeRule `json:"minStopOrProfitDistance,omitempty"`
}

// MarketSnapshot holds current bid/offer.
type MarketSnapshot struct {
	Bid float64 `json:"bid"`
	Offer float64 `json:"offer"`
	MarketStatus string `json:"marketStatus"`
}

// SearchMarkets returns Capital.com market search results for a term.
func (c *Client) SearchMarkets(ctx context.Context, searchTerm string) ([]MarketInfo, error) {
	path := fmt.Sprintf("/api/v1/markets?searchTerm=%s", searchTerm)
	data, err := c.Do(ctx, "GET", path, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("search markets: %w", err)
	}
	var mr MarketsResponse
	if err := json.Unmarshal(data, &mr); err != nil {
		return nil, fmt.Errorf("parse markets: %w", err)
	}
	return mr.Markets, nil
}

// PickTradeableCanary selects a TRADEABLE market with usable dealing-rule hints.
// It never silently returns markets[0] unless that row is the only TRADEABLE match.
func PickTradeableCanary(markets []MarketInfo) (MarketInfo, error) {
	var ok []MarketInfo
	for _, m := range markets {
		if !strings.EqualFold(strings.TrimSpace(m.MarketStatus), "TRADEABLE") {
			continue
		}
		if strings.TrimSpace(m.Epic) == "" {
			continue
		}
		ok = append(ok, m)
	}
	if len(ok) == 0 {
		return MarketInfo{}, fmt.Errorf("no TRADEABLE market in search results")
	}
	if len(ok) == 1 {
		return ok[0], nil
	}
	for _, m := range ok {
		if strings.EqualFold(m.InstrumentType, "CRYPTOCURRENCIES") || strings.Contains(strings.ToUpper(m.InstrumentType), "CRYPTO") {
			return m, nil
		}
	}
	return ok[0], nil
}

// ResolveEpic finds the XAUUSD/gold epic via search. Returns the first COMMODITIES match or first match.
func (c *Client) ResolveEpic(ctx context.Context, searchTerm string) (string, error) {
	path := fmt.Sprintf("/api/v1/markets?searchTerm=%s", searchTerm)
	data, err := c.Do(ctx, "GET", path, nil, nil)
	if err != nil {
		return "", fmt.Errorf("resolve epic: %w", err)
	}
	var mr MarketsResponse
	if err := json.Unmarshal(data, &mr); err != nil {
		return "", fmt.Errorf("parse markets: %w", err)
	}
	if len(mr.Markets) == 0 {
		return "", fmt.Errorf("no markets found for search %q", searchTerm)
	}
	for _, m := range mr.Markets {
		if m.InstrumentType == "COMMODITIES" {
			return m.Epic, nil
		}
	}
	return mr.Markets[0].Epic, nil
}

// GetMarketDetails returns dealing rules and snapshot for an epic (optional; use after session).
func (c *Client) GetMarketDetails(ctx context.Context, epic string) (*MarketDetailsResponse, error) {
	path := fmt.Sprintf("/api/v1/markets/%s", epic)
	data, err := c.Do(ctx, "GET", path, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("get market details: %w", err)
	}
	var md MarketDetailsResponse
	if err := json.Unmarshal(data, &md); err != nil {
		return nil, fmt.Errorf("parse market details: %w", err)
	}
	return &md, nil
}
