package market

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	Epic       string  `json:"epic,omitempty"`
	ProfitLoss  float64 `json:"profitLoss,omitempty"`
	Upnl        float64 `json:"upl,omitempty"`
	StopLevel   float64 `json:"stopLevel,omitempty"`
	ProfitLevel float64 `json:"profitLevel,omitempty"`
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

// GetPosition returns a single open position by dealId.
func (c *Client) GetPosition(ctx context.Context, dealID string) (*PositionItem, error) {
	if strings.TrimSpace(dealID) == "" {
		return nil, fmt.Errorf("get position: empty dealId")
	}
	data, err := c.Do(ctx, "GET", "/api/v1/positions/"+dealID, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("get position: %w", err)
	}
	var item PositionItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, fmt.Errorf("parse position: %w", err)
	}
	return &item, nil
}

// DealAck is the broker acknowledgement (dealReference) for open/close/update.
type DealAck struct {
	DealReference string `json:"dealReference"`
}

// ClosePosition sends DELETE /api/v1/positions/{dealId}.
func (c *Client) ClosePosition(ctx context.Context, dealID string) (*DealAck, error) {
	if strings.TrimSpace(dealID) == "" {
		return nil, fmt.Errorf("close position: empty dealId")
	}
	data, err := c.Do(ctx, "DELETE", "/api/v1/positions/"+dealID, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("close position: %w", err)
	}
	var ack DealAck
	if err := json.Unmarshal(data, &ack); err != nil {
		return nil, fmt.Errorf("parse close position: %w", err)
	}
	return &ack, nil
}

// UpdatePositionRequest is the body for PUT /api/v1/positions/{dealId}.
type UpdatePositionRequest struct {
	StopLevel    *float64 `json:"stopLevel,omitempty"`
	ProfitLevel  *float64 `json:"profitLevel,omitempty"`
	TrailingStop *bool    `json:"trailingStop,omitempty"`
}

// UpdatePosition sends PUT /api/v1/positions/{dealId}.
func (c *Client) UpdatePosition(ctx context.Context, dealID string, req UpdatePositionRequest) (*DealAck, error) {
	if strings.TrimSpace(dealID) == "" {
		return nil, fmt.Errorf("update position: empty dealId")
	}
	data, err := c.Do(ctx, "PUT", "/api/v1/positions/"+dealID, req, nil)
	if err != nil {
		return nil, fmt.Errorf("update position: %w", err)
	}
	var ack DealAck
	if err := json.Unmarshal(data, &ack); err != nil {
		return nil, fmt.Errorf("parse update position: %w", err)
	}
	return &ack, nil
}

// ConfirmDeal fetches GET /api/v1/confirms/{dealReference}.
func (c *Client) ConfirmDeal(ctx context.Context, dealReference string) (*ConfirmDealResponse, error) {
	if strings.TrimSpace(dealReference) == "" {
		return nil, fmt.Errorf("confirm deal: empty dealReference")
	}
	data, err := c.Do(ctx, "GET", "/api/v1/confirms/"+dealReference, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("confirm deal: %w", err)
	}
	var cr ConfirmDealResponse
	if err := json.Unmarshal(data, &cr); err != nil {
		return nil, fmt.Errorf("parse confirm: %w", err)
	}
	return &cr, nil
}

// ConfirmDealResponse is the Capital.com deal confirmation payload.
type ConfirmDealResponse struct {
	DealReference string  `json:"dealReference"`
	DealID        string  `json:"dealId"`
	Status        string  `json:"status"`
	DealStatus    string  `json:"dealStatus"`
	Level         float64 `json:"level"`
	Size          float64 `json:"size"`
	Direction     string  `json:"direction"`
	Epic          string  `json:"epic"`
	Reason        string  `json:"reason,omitempty"`
	ProfitLoss    float64 `json:"profitLoss,omitempty"`
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
