package telemetry

import (
	"context"
	"errors"
	"strings"
	"time"

	"aurumflow/config"
	"aurumflow/internal/logger"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/db"
	"google.golang.org/api/option"
)

var (
	errInvalidCredentials = errors.New("invalid or missing credentials (need project_id, client_email, private_key)")
	errNoCredentials       = errors.New("no credentials: set FIREBASE_SERVICE_ACCOUNT_JSON, GOOGLE_APPLICATION_CREDENTIALS, or service_account_path")
)

const (
	rtdbLivePrefix   = "v1/live/instances"
	schemaVersion    = 1
	analyticsVersion = 1
	publishTimeout   = 5 * time.Second
)

// NormalizeInstanceID replaces characters invalid in RTDB keys (. $ # [ ] /) so the path is valid.
func NormalizeInstanceID(instanceID string) string {
	s := strings.TrimSpace(instanceID)
	if s == "" {
		return "default"
	}
	replacer := strings.NewReplacer(
		".", "_", "$", "_", "#", "_", "[", "_", "]", "_", "/", "_",
	)
	return replacer.Replace(s)
}

// FirebasePublisher implements LivePublisher using Firebase Realtime Database.
type FirebasePublisher struct {
	client *db.Client
}

// NewLivePublisher creates a LivePublisher for Firebase RTDB. Returns nil, nil if config is disabled or invalid.
func NewLivePublisher(ctx context.Context, cfg *config.FirebaseTelemetryConfig) (LivePublisher, error) {
	if cfg == nil || !cfg.Enabled {
		return nil, nil
	}
	rtdbURL := strings.TrimSpace(cfg.RTDBURL)
	if rtdbURL == "" {
		return nil, nil
	}
	jsonBytes, filePath, err := resolveCredentials(cfg.ServiceAccountPath)
	if err != nil {
		return nil, err
	}
	var opts []option.ClientOption
	if len(jsonBytes) > 0 {
		if !validateCredentials(jsonBytes) {
			return nil, errInvalidCredentials
		}
		opts = append(opts, option.WithCredentialsJSON(jsonBytes))
	} else if filePath != "" {
		opts = append(opts, option.WithCredentialsFile(filePath))
	} else {
		return nil, errNoCredentials
	}
	app, err := firebase.NewApp(ctx, &firebase.Config{DatabaseURL: rtdbURL}, opts...)
	if err != nil {
		return nil, err
	}
	client, err := app.DatabaseWithURL(ctx, rtdbURL)
	if err != nil {
		return nil, err
	}
	return &FirebasePublisher{client: client}, nil
}

// PublishMeta writes run metadata once at start.
func (p *FirebasePublisher) PublishMeta(ctx context.Context, instanceID, runID string, meta MetaPayload) error {
	ctx, cancel := context.WithTimeout(ctx, publishTimeout)
	defer cancel()
	path := pathFor(instanceID, "meta")
	ref := p.client.NewRef(path)
	if meta.SchemaVersion == 0 {
		meta.SchemaVersion = schemaVersion
	}
	if meta.AnalyticsVersion == 0 {
		meta.AnalyticsVersion = analyticsVersion
	}
	return ref.Set(ctx, meta)
}

// PublishStatus writes heartbeat/status snapshot.
func (p *FirebasePublisher) PublishStatus(ctx context.Context, instanceID, runID string, status StatusSnapshot) error {
	ctx, cancel := context.WithTimeout(ctx, publishTimeout)
	defer cancel()
	path := pathFor(instanceID, "status")
	ref := p.client.NewRef(path)
	if status.SchemaVersion == 0 {
		status.SchemaVersion = schemaVersion
	}
	return ref.Set(ctx, status)
}

// PublishPositions writes open positions summary (MVP: openCount, maxTrades, lastDealRef).
func (p *FirebasePublisher) PublishPositions(ctx context.Context, instanceID string, openCount, maxTrades int, lastDealRef string) error {
	ctx, cancel := context.WithTimeout(ctx, publishTimeout)
	defer cancel()
	path := pathFor(instanceID, "positions/open")
	ref := p.client.NewRef(path)
	payload := map[string]interface{}{
		"openCount":   openCount,
		"maxTrades":   maxTrades,
		"lastDealRef": lastDealRef,
		"updatedAt":   time.Now().UTC().Format(time.RFC3339),
	}
	return ref.Set(ctx, payload)
}

func pathFor(instanceID, suffix string) string {
	return rtdbLivePrefix + "/" + NormalizeInstanceID(instanceID) + "/" + suffix
}

// PublishStatusBestEffort calls PublishStatus and logs a warning on error (best-effort).
func PublishStatusBestEffort(p LivePublisher, ctx context.Context, instanceID, runID string, status StatusSnapshot) {
	if p == nil {
		return
	}
	if err := p.PublishStatus(ctx, instanceID, runID, status); err != nil {
		logger.Warn("telemetry publish status failed: %v", err)
	}
}

// PublishPositionsBestEffort calls PublishPositions and logs a warning on error (best-effort).
func PublishPositionsBestEffort(p LivePublisher, ctx context.Context, instanceID string, openCount, maxTrades int, lastDealRef string) {
	if p == nil {
		return
	}
	if err := p.PublishPositions(ctx, instanceID, openCount, maxTrades, lastDealRef); err != nil {
		logger.Warn("telemetry publish positions failed: %v", err)
	}
}
