package notifications

import (
	"context"
	"strings"
	"testing"

	"aurumflow/config"
)

func TestEscapeHTMLValue(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"plain", "plain"},
		{"a & b", "a &amp; b"},
		{"x < y", "x &lt; y"},
		{"c > d", "c &gt; d"},
		{"<b>tag</b>", "&lt;b&gt;tag&lt;/b&gt;"},
		{"foo &amp; bar", "foo &amp;amp; bar"},
	}
	for _, tt := range tests {
		got := EscapeHTMLValue(tt.in)
		if got != tt.want {
			t.Errorf("EscapeHTMLValue(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestTruncateForTelegram(t *testing.T) {
	short := "hello"
	if got := truncateForTelegram(short); got != short {
		t.Errorf("truncateForTelegram(short) = %q, want %q", got, short)
	}
	long := strings.Repeat("x", 5000)
	got := truncateForTelegram(long)
	if len(got) > 4096 {
		t.Errorf("truncateForTelegram(long) length %d > 4096", len(got))
	}
	if !strings.HasSuffix(got, telegramTruncateSuffix) {
		t.Errorf("truncateForTelegram(long) should end with %q, got %q", telegramTruncateSuffix, got[len(got)-20:])
	}
	wantMax := telegramTruncateAt + len(telegramTruncateSuffix)
	if len(got) > wantMax {
		t.Errorf("truncateForTelegram(long) length %d > %d", len(got), wantMax)
	}
}

func TestNewTelegramClient_NilConfig(t *testing.T) {
	c := NewTelegramClient(nil)
	if c != nil {
		t.Error("NewTelegramClient(nil) should return nil")
	}
}

func TestNewTelegramClient_EmptyCredentials(t *testing.T) {
	c := NewTelegramClient(&config.TelegramConfig{Enabled: true, BotToken: "", ChatID: "123"})
	if c != nil {
		t.Error("NewTelegramClient with empty token should return nil")
	}
	c = NewTelegramClient(&config.TelegramConfig{Enabled: true, BotToken: "token", ChatID: ""})
	if c != nil {
		t.Error("NewTelegramClient with empty chat_id should return nil")
	}
}

func TestTelegramClient_Send_EmptyText(t *testing.T) {
	// Send with empty text should not panic; client is nil so returns nil
	var c *TelegramClient
	if err := c.Send(context.Background(), ""); err != nil {
		t.Errorf("Send with nil client and empty text: %v", err)
	}
}
