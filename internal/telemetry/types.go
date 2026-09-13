package telemetry

import "context"

// LivePublisher publishes live telemetry to a backend (e.g. Firebase RTDB).
// Implementations should be best-effort: log errors but do not fail the loop.
type LivePublisher interface {
	PublishMeta(ctx context.Context, instanceID, runID string, meta MetaPayload) error
	PublishStatus(ctx context.Context, instanceID, runID string, status StatusSnapshot) error
	PublishPositions(ctx context.Context, instanceID string, openCount, maxTrades int, lastDealRef string) error
}

// MetaPayload holds run metadata written once at start.
type MetaPayload struct {
	RunID            string `json:"runId"`
	Epic             string `json:"epic"`
	StartedAt        string `json:"startedAt"`
	SchemaVersion    int    `json:"schemaVersion"`
	AnalyticsVersion int    `json:"analyticsVersion,omitempty"`
	BotTag           string `json:"botTag,omitempty"`
	BotTagNormalized string `json:"botTagNormalized,omitempty"`
}

// StatusSnapshot holds heartbeat/status data written on each heartbeat.
type StatusSnapshot struct {
	State            string  `json:"state"`
	Balance          float64 `json:"balance"`
	OpenPositions    int     `json:"openPositions"`
	LastDecision     string  `json:"lastDecision"`
	MarketStatus     string  `json:"marketStatus"`
	DailyDrawdownPct float64 `json:"dailyDrawdownPct"`
	Tick             int     `json:"tick"`
	UpdatedAt        string  `json:"updatedAt"`
	RunID            string  `json:"runId,omitempty"`
	SchemaVersion    int     `json:"schemaVersion,omitempty"`
}
