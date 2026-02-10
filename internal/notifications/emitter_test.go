package notifications

import (
	"testing"

	"aurumflow/config"
)

func TestNewEmitter_NilConfig(t *testing.T) {
	e := NewEmitter(nil, nil)
	if e == nil {
		t.Fatal("NewEmitter(nil) should not return nil")
	}
	if e.Active() {
		t.Error("emitter with nil config should not be active")
	}
	e.Emit(NotifEvent{Type: TypeHeartbeat}) // should not panic
}

func TestNewEmitter_Disabled(t *testing.T) {
	cfg := &config.Config{
		Notifications: &config.NotificationsConfig{
			Enabled:  false,
			Provider: "pushover",
		},
	}
	e := NewEmitter(cfg, nil)
	if e.Active() {
		t.Error("emitter with enabled=false should not be active")
	}
}

func TestNewEmitter_NoCredentials(t *testing.T) {
	cfg := &config.Config{
		API: config.APIConfig{Mode: config.ModeDemo},
		Notifications: &config.NotificationsConfig{
			Enabled:  true,
			Provider: "pushover",
			Pushover: &config.PushoverConfig{Token: "", User: ""},
		},
	}
	e := NewEmitter(cfg, nil)
	if e.Active() {
		t.Error("emitter with empty token/user should not be active")
	}
}

func TestEmitter_EmitNoOp(t *testing.T) {
	e := NewEmitter(nil, nil)
	e.Emit(NotifEvent{Type: TypeBotStart, Category: CategoryHealth})
	e.Emit(NotifEvent{Type: TypeSignalGenerated})
	// no panic, no send
}

func TestNewEmitter_TelegramOnly(t *testing.T) {
	cfg := &config.Config{
		API: config.APIConfig{Mode: config.ModeDemo},
		Notifications: &config.NotificationsConfig{
			Enabled:  true,
			Provider: "pushover",
			Pushover: &config.PushoverConfig{Token: "", User: ""},
			Telegram: &config.TelegramConfig{
				Enabled:  true,
				BotToken: "fake-token",
				ChatID:   "-100123",
				ParseMode: "HTML",
			},
		},
	}
	e := NewEmitter(cfg, nil)
	if e == nil {
		t.Fatal("NewEmitter should not return nil")
	}
	if !e.Active() {
		t.Error("emitter with only Telegram configured should be active")
	}
}

func TestNewEmitter_TelegramEnabledButNoCredentials(t *testing.T) {
	cfg := &config.Config{
		API: config.APIConfig{Mode: config.ModeDemo},
		Notifications: &config.NotificationsConfig{
			Enabled:  true,
			Provider: "pushover",
			Pushover: &config.PushoverConfig{Token: "", User: ""},
			Telegram: &config.TelegramConfig{
				Enabled:  true,
				BotToken: "",
				ChatID:   "",
			},
		},
	}
	e := NewEmitter(cfg, nil)
	if e == nil {
		t.Fatal("NewEmitter should not return nil")
	}
	if e.Active() {
		t.Error("emitter with Telegram enabled but missing bot_token/chat_id should not be active")
	}
}
