package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"aurumflow/config"
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
	Reason        string   `json:"reason,omitempty"`
	ProfitLoss    float64  `json:"profitLoss,omitempty"`
}

func (e *Executor) Mode() config.ExecutionMode { return config.ExecutionDemo }

func (e *Executor) throttle() {
	e.mu.Lock()
	now := time.Now()
	if now.Before(e.rateLimit) {
		time.Sleep(time.Until(e.rateLimit))
	}
	e.rateLimit = time.Now().Add(100 * time.Millisecond)
	e.mu.Unlock()
}

// OpenInfrastructure submits a single DEMO canary order with no strategy SL/TP.
func (e *Executor) OpenInfrastructure(ctx context.Context, direction string, size float64) (*OpenResult, error) {
	e.throttle()
	req := CreatePositionRequest{
		Epic:           e.Epic,
		Direction:      direction,
		Size:           size,
		GuaranteedStop: false,
	}
	data, err := e.Client.Do(ctx, "POST", "/api/v1/positions", req, nil)
	if err != nil {
		return nil, err
	}
	var resp CreatePositionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse position response: %w", err)
	}
	if strings.TrimSpace(resp.DealReference) == "" {
		return nil, fmt.Errorf("open infrastructure: empty dealReference")
	}
	return &OpenResult{DealReference: resp.DealReference}, nil
}

func (e *Executor) OpenPosition(ctx context.Context, req OpenRequest) (*OpenResult, error) {
	sig := req.Signal
	if sig == nil {
		sig = &models.TradeSignal{Direction: req.Direction, StopLoss: req.StopLoss, TakeProfit: req.TakeProfit}
	}
	ref, err := e.SendOrder(ctx, sig, req.Size)
	if err != nil {
		return nil, err
	}
	return &OpenResult{DealReference: ref}, nil
}

func (e *Executor) Confirm(ctx context.Context, dealReference string) (*ConfirmResult, error) {
	cr, err := e.ConfirmDeal(ctx, dealReference)
	if err != nil {
		return nil, err
	}
	return &ConfirmResult{
		DealReference: cr.DealReference,
		DealID:        cr.DealID,
		Status:        cr.Status,
		DealStatus:    cr.DealStatus,
		Level:         cr.Level,
		Size:          cr.Size,
		Direction:     cr.Direction,
		Epic:          cr.Epic,
		Reason:        cr.Reason,
		ProfitLoss:    cr.ProfitLoss,
	}, nil
}

func (e *Executor) ClosePosition(ctx context.Context, dealID string) (*CloseResult, error) {
	e.throttle()
	ack, err := e.Client.ClosePosition(ctx, dealID)
	if err != nil {
		return nil, err
	}
	out := &CloseResult{DealReference: ack.DealReference, DealID: dealID}
	if ack.DealReference != "" {
		if cr, cerr := e.ConfirmDeal(ctx, ack.DealReference); cerr == nil && cr != nil {
			out.Status = cr.Status
			out.Level = cr.Level
			out.PnL = cr.ProfitLoss
			if cr.DealID != "" {
				out.DealID = cr.DealID
			}
		}
	}
	return out, nil
}

func (e *Executor) UpdatePosition(ctx context.Context, req UpdateRequest) (*CloseResult, error) {
	e.throttle()
	ack, err := e.Client.UpdatePosition(ctx, req.DealID, market.UpdatePositionRequest{
		StopLevel:    req.StopLevel,
		ProfitLevel:  req.ProfitLevel,
		TrailingStop: req.TrailingStop,
	})
	if err != nil {
		return nil, err
	}
	out := &CloseResult{DealReference: ack.DealReference, DealID: req.DealID}
	if ack.DealReference != "" {
		if cr, cerr := e.ConfirmDeal(ctx, ack.DealReference); cerr == nil && cr != nil {
			out.Status = cr.Status
			out.Level = cr.Level
		}
	}
	return out, nil
}

func (e *Executor) Positions(ctx context.Context) ([]Position, error) {
	pr, err := e.Client.GetPositions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Position, 0, len(pr.Positions))
	for _, p := range pr.Positions {
		out = append(out, Position{
			DealID:    p.Position.DealID,
			Epic:      p.GetEpic(),
			Direction: p.Position.Direction,
			Size:      p.Position.Size,
			Level:     p.Position.Level,
		})
	}
	return out, nil
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
	md, err := e.Client.ConfirmDeal(ctx, dealReference)
	if err != nil {
		return nil, err
	}
	cr := &ConfirmResponse{
		DealReference: md.DealReference,
		DealID:        md.DealID,
		Status:        md.Status,
		DealStatus:    md.DealStatus,
		Level:         md.Level,
		Size:          md.Size,
		Direction:     md.Direction,
		Epic:          md.Epic,
		Reason:        md.Reason,
		ProfitLoss:    md.ProfitLoss,
	}
	logger.Info("execution: confirm dealRef=%s status=%s dealStatus=%s dealId=%s",
		cr.DealReference, cr.Status, cr.DealStatus, cr.DealID)
	return cr, nil
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
