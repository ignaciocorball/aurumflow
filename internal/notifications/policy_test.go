package notifications

import (
	"testing"
	"time"

	"aurumflow/config"
)

func TestNewPolicy_NilConfig(t *testing.T) {
	p := NewPolicy(nil)
	if p == nil {
		t.Fatal("NewPolicy(nil) should not return nil")
	}
	if p.CategoryEnabled(CategoryHealth) {
		t.Error("CategoryEnabled with nil config should be false")
	}
	if p.ShouldNotify(NotifEvent{Category: CategoryHealth}) {
		t.Error("ShouldNotify with nil config should be false")
	}
}

func TestPolicy_WithConfig(t *testing.T) {
	cfg := &config.NotificationsConfig{
		Enabled:  true,
		Provider: "pushover",
		Categories: &config.NotifCategoriesConfig{
			Health: true, Market: false, Signals: true,
		},
		RateLimit: &config.NotifRateLimitConfig{
			DedupWindowSeconds:      900,
			AggregateWindowSeconds: 1800,
		},
	}
	p := NewPolicy(cfg)
	if !p.CategoryEnabled(CategoryHealth) {
		t.Error("CategoryHealth should be enabled")
	}
	if p.CategoryEnabled(CategoryMarket) {
		t.Error("CategoryMarket should be disabled")
	}
	if p.MinInterval(TypeHeartbeat) != 0 {
		t.Error("MinInterval without map should be 0")
	}
	cfg.RateLimit.MinIntervalSecondsByEvent = map[string]int{"HEARTBEAT": 3600}
	if p.MinInterval(TypeHeartbeat) != time.Hour {
		t.Errorf("MinInterval(HEARTBEAT) = %v, want 1h", p.MinInterval(TypeHeartbeat))
	}
	if p.DedupWindow() != 900*time.Second {
		t.Errorf("DedupWindow = %v, want 900s", p.DedupWindow())
	}
}

func TestPolicy_DedupKey(t *testing.T) {
	p := NewPolicy(&config.NotificationsConfig{})
	ev := NotifEvent{Type: TypeSignalRejected, Instrument: "ETHUSD", Session: "NY", Payload: map[string]any{"reason_code": RejH1Filter}}
	key := p.DedupKey(ev)
	if key == "" {
		t.Error("DedupKey should not be empty")
	}
	ev2 := NotifEvent{Type: TypeSignalRejected, Instrument: "ETHUSD", Session: "NY", Payload: map[string]any{"reason_code": RejH1Filter}}
	if p.DedupKey(ev2) != key {
		t.Error("same event should produce same DedupKey")
	}
}

func TestPolicy_SweepConfirmedShouldNotify(t *testing.T) {
	cfg := &config.NotificationsConfig{
		Categories: &config.NotifCategoriesConfig{Market: true},
		Thresholds: &config.NotifThresholdsConfig{SweepMinStrength: 0.5},
	}
	p := NewPolicy(cfg)
	ev := NotifEvent{Category: CategoryMarket, Payload: map[string]any{"strength": 0.6, "zoneBuy": 0, "zoneSell": 0}}
	if !p.SweepConfirmedShouldNotify(ev) {
		t.Error("strength >= min should notify")
	}
	ev.Payload["strength"] = 0.1
	ev.Payload["zoneBuy"] = 100.0
	if !p.SweepConfirmedShouldNotify(ev) {
		t.Error("zoneBuy != 0 should notify")
	}
	ev.Payload["zoneBuy"] = 0
	ev.Payload["zoneSell"] = 0
	ev.Payload["typeChanged"] = true
	if !p.SweepConfirmedShouldNotify(ev) {
		t.Error("typeChanged should notify")
	}
	ev.Payload["typeChanged"] = false
	if p.SweepConfirmedShouldNotify(ev) {
		t.Error("low strength and no zone/change should not notify")
	}
}
