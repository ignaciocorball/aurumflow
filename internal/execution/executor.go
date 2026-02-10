package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"aurumflow/internal/logger"
	"aurumflow/internal/market"
	"aurumflow/pkg/models"
)

// Executor sends orders to Capital.com and confirms them.
type Executor struct {
	Client    *market.Client
	Epic      string
	MinSize   float64
	SizeStep  float64
	rateLimit time.Time
	mu        sync.Mutex
}

// NewExecutor creates an execution layer with rate limit (100 ms between POST positions).
func NewExecutor(client *market.Client, epic string, minSize, sizeStep float64) *Executor {
	if minSize <= 0 {
		minSize = 0.1
	}
	if sizeStep <= 0 {
		sizeStep = 0.1
	}
	return &Executor{
		Client:   client,
		Epic:     epic,
		MinSize:  minSize,
		SizeStep: sizeStep,
	}
}

// CreatePositionRequest is the body for POST /api/v1/positions.
type CreatePositionRequest struct {
	Epic         string  `json:"epic"`
	Direction    string  `json:"direction"`
	Size         float64 `json:"size"`
	StopLevel    *float64 `json:"stopLevel,omitempty"`
	StopDistance *float64 `json:"stopDistance,omitempty"`
	ProfitLevel  *float64 `json:"profitLevel,omitempty"`
	ProfitDistance *float64 `json:"profitDistance,omitempty"`
	GuaranteedStop bool   `json:"guaranteedStop"`
}

// CreatePositionResponse holds dealReference from API.
type CreatePositionResponse struct {
	DealReference string `json:"dealReference"`
}

// ConfirmResponse is the response from GET /api/v1/confirms/{dealReference}.
type ConfirmResponse struct {
	DealReference string   `json:"dealReference"`
	DealID        string   `json:"dealId"`
	Status        string   `json:"status"`
	DealStatus    string   `json:"dealStatus"`
	Level         float64  `json:"level"`
	Size          float64  `json:"size"`
	Direction     string   `json:"direction"`
	Epic          string   `json:"epic"`
}

// SendOrder builds the position request, enforces 100 ms rate limit, sends POST, then confirms.
func (e *Executor) SendOrder(ctx context.Context, signal *models.TradeSignal, size float64) (dealRef string, err error) {
	if signal == nil {
		return "", fmt.Errorf("nil signal")
	}
	e.mu.Lock()
	now := time.Now()
	if now.Before(e.rateLimit) {
		time.Sleep(time.Until(e.rateLimit))
	}
	e.rateLimit = time.Now().Add(100 * time.Millisecond)
	e.mu.Unlock()

	stopDist := signal.Entry - signal.StopLoss
	if stopDist < 0 {
		stopDist = -stopDist
	}
	profitDist := signal.TakeProfit - signal.Entry
	if profitDist < 0 {
		profitDist = -profitDist
	}
	req := CreatePositionRequest{
		Epic:            e.Epic,
		Direction:       signal.Direction,
		Size:            size,
		StopLevel:       &signal.StopLoss,
		ProfitLevel:     &signal.TakeProfit,
		GuaranteedStop:  false,
	}
	data, err := e.Client.Do(ctx, "POST", "/api/v1/positions", req, nil)
	if err != nil {
		logger.Error("execution: POST positions failed: %v", err)
		return "", err
	}
	var resp CreatePositionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("parse position response: %w", err)
	}
	dealRef = resp.DealReference
	logger.Info("execution: position sent dealRef=%s size=%.2f direction=%s entry=%.2f sl=%.2f tp=%.2f",
		dealRef, size, signal.Direction, signal.Entry, signal.StopLoss, signal.TakeProfit)

	_ = stopDist
	_ = profitDist
	return dealRef, nil
}

// ConfirmDeal polls GET /api/v1/confirms/{dealReference} and logs result.
func (e *Executor) ConfirmDeal(ctx context.Context, dealReference string) (*ConfirmResponse, error) {
	path := "/api/v1/confirms/" + dealReference
	data, err := e.Client.Do(ctx, "GET", path, nil, nil)
	if err != nil {
		return nil, err
	}
	var cr ConfirmResponse
	if err := json.Unmarshal(data, &cr); err != nil {
		return nil, fmt.Errorf("parse confirm: %w", err)
	}
	logger.Info("execution: confirm dealRef=%s status=%s dealStatus=%s dealId=%s",
		cr.DealReference, cr.Status, cr.DealStatus, cr.DealID)
	return &cr, nil
}

// LogTrade writes a trade log line to terminal (structured for future metrics).
func LogTrade(signal *models.TradeSignal, size float64, dealRef string, err error) {
	if err != nil {
		logger.Error("trade FAIL direction=%s entry=%.2f size=%.2f dealRef=%s err=%v",
			signal.Direction, signal.Entry, size, dealRef, err)
		return
	}
	logger.Info("trade OK direction=%s entry=%.2f sl=%.2f tp=%.2f size=%.2f score=%d dealRef=%s",
		signal.Direction, signal.Entry, signal.StopLoss, signal.TakeProfit, size, signal.Score, dealRef)
}
