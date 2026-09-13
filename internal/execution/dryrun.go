package execution

import (
	"context"
	"fmt"

	"aurumflow/config"
)

// DryRunProvider simulates execution and never calls mutating HTTP methods.
type DryRunProvider struct {
	Inner ExecutionProvider // used only for read-only Positions
}

func (p *DryRunProvider) Mode() config.ExecutionMode { return config.ExecutionDryRun }

func (p *DryRunProvider) OpenPosition(ctx context.Context, req OpenRequest) (*OpenResult, error) {
	entry := req.StopLoss
	_ = entry
	if req.Signal != nil {
		entry = req.Signal.Entry
	}
	return nil, fmt.Errorf("%w: DRY_RUN would_have_sent epic=%s direction=%s size=%.4f entry=%.4f sl=%.4f tp=%.4f",
		ErrMutationDisabled, req.Epic, req.Direction, req.Size, entry, req.StopLoss, req.TakeProfit)
}

func (p *DryRunProvider) Confirm(ctx context.Context, dealReference string) (*ConfirmResult, error) {
	return nil, fmt.Errorf("%w: DRY_RUN confirm skipped", ErrMutationDisabled)
}

func (p *DryRunProvider) ClosePosition(ctx context.Context, dealID string) (*CloseResult, error) {
	return nil, fmt.Errorf("%w: DRY_RUN close skipped", ErrMutationDisabled)
}

func (p *DryRunProvider) UpdatePosition(ctx context.Context, req UpdateRequest) (*CloseResult, error) {
	return nil, fmt.Errorf("%w: DRY_RUN update skipped", ErrMutationDisabled)
}

func (p *DryRunProvider) Positions(ctx context.Context) ([]Position, error) {
	if p.Inner != nil {
		return p.Inner.Positions(ctx)
	}
	return nil, nil
}
