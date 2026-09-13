package execution

import (
	"context"
	"fmt"

	"aurumflow/config"
	"aurumflow/pkg/models"
)

// OpenRequest is a broker-agnostic open.
type OpenRequest struct {
	Epic, Direction string
	Size            float64
	StopLoss        float64
	TakeProfit      float64
	Signal          *models.TradeSignal
}

// OpenResult is the broker ack after submit.
type OpenResult struct {
	DealReference string
}

// ConfirmResult is a deal confirmation.
type ConfirmResult struct {
	DealReference string
	DealID        string
	Status        string
	DealStatus    string
	Level         float64
	Size          float64
	Direction     string
	Epic          string
}

// CloseResult is a close or update acknowledgement plus optional confirm.
type CloseResult struct {
	DealReference string
	DealID        string
	Status        string
	Level         float64
}

// UpdateRequest updates SL/TP on an open position.
type UpdateRequest struct {
	DealID      string
	StopLevel   *float64
	ProfitLevel *float64
	TrailingStop *bool
}

// Position is a normalized open position.
type Position struct {
	DealID    string
	Epic      string
	Direction string
	Size      float64
	Level     float64
}

// ExecutionProvider is the small P1 broker mutation surface.
type ExecutionProvider interface {
	Mode() config.ExecutionMode
	OpenPosition(ctx context.Context, req OpenRequest) (*OpenResult, error)
	Confirm(ctx context.Context, dealReference string) (*ConfirmResult, error)
	ClosePosition(ctx context.Context, dealID string) (*CloseResult, error)
	UpdatePosition(ctx context.Context, req UpdateRequest) (*CloseResult, error)
	Positions(ctx context.Context) ([]Position, error)
}

// ErrMutationDisabled is returned when DISABLED/DRY_RUN would mutate the broker.
var ErrMutationDisabled = fmt.Errorf("execution mutation blocked")
