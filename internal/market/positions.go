package market

import (
	"context"
	"encoding/json"
	"fmt"
)

// PositionsResponse is the response from GET /api/v1/positions.
type PositionsResponse struct {
	Positions []PositionItem `json:"positions"`
}

// PositionItem wraps position and market.
type PositionItem struct {
	Position PositionData `json:"position"`
	Market   *MarketInfo  `json:"market,omitempty"`
}

// PositionData holds dealId, size, direction, level, and optionally epic.
type PositionData struct {
	DealID   string  `json:"dealId"`
	Size     float64 `json:"size"`
	Direction string `json:"direction"`
	Level   float64 `json:"level"`
	Epic    string  `json:"epic,omitempty"`
}

// GetEpic returns the epic for this position (from Market.Epic or Position.Epic).
func (p *PositionItem) GetEpic() string {
	if p.Market != nil && p.Market.Epic != "" {
		return p.Market.Epic
	}
	if p.Position.Epic != "" {
		return p.Position.Epic
	}
	return ""
}

// GetPositions returns all open positions for the active account.
func (c *Client) GetPositions(ctx context.Context) (*PositionsResponse, error) {
	data, err := c.Do(ctx, "GET", "/api/v1/positions", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("get positions: %w", err)
	}
	var pr PositionsResponse
	if err := json.Unmarshal(data, &pr); err != nil {
		return nil, fmt.Errorf("parse positions: %w", err)
	}
	return &pr, nil
}
