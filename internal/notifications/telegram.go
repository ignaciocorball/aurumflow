package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"aurumflow/config"
)

const (
	telegramAPIBase     = "https://api.telegram.org/bot"
	telegramSendTimeout = 10 * time.Second
	telegramMaxLen      = 4096
	telegramTruncateAt  = 3800
	telegramTruncateSuffix = " …(truncado)"
)

// TelegramClient sends messages to the Telegram Bot API.
type TelegramClient struct {
	baseURL   string
	botToken  string
	chatID    string
	parseMode string
	client    *http.Client
}

// NewTelegramClient builds a client from config. Returns nil if cfg is nil or token/chatID empty.
func NewTelegramClient(cfg *config.TelegramConfig) *TelegramClient {
	if cfg == nil {
		return nil
	}
	token := strings.TrimSpace(cfg.BotToken)
	chatID := strings.TrimSpace(cfg.ChatID)
	if token == "" || chatID == "" {
		return nil
	}
	parseMode := strings.TrimSpace(cfg.ParseMode)
	if parseMode == "" {
		parseMode = "HTML"
	}
	return &TelegramClient{
		baseURL:   telegramAPIBase + token + "/",
		botToken:  token,
		chatID:    chatID,
		parseMode: parseMode,
		client:    &http.Client{Timeout: telegramSendTimeout},
	}
}

// EscapeHTMLValue escapes & < > in variable values for safe use inside HTML tags.
// Do not use on markup; only on dynamic values (symbols, numbers, reasons, etc.).
func EscapeHTMLValue(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// truncateForTelegram ensures text is at most telegramMaxLen; truncates at telegramTruncateAt and appends suffix.
func truncateForTelegram(text string) string {
	if len(text) <= telegramMaxLen {
		return text
	}
	// Truncate so that result + suffix fits; suffix may be multi-byte
	maxBody := telegramTruncateAt
	if maxBody+len(telegramTruncateSuffix) > telegramMaxLen {
		maxBody = telegramMaxLen - len(telegramTruncateSuffix)
	}
	return text[:maxBody] + telegramTruncateSuffix
}

// sendMessageRequest is the JSON body for sendMessage.
type sendMessageRequest struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// sendMessageResponse is the minimal response from Telegram API.
type sendMessageResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description,omitempty"`
}

// Send posts a message to the Telegram channel/chat. Best-effort: one POST, no retries.
// Truncates text to 3800 chars + " …(truncado)" if over 4096. Does not log token.
func (c *TelegramClient) Send(ctx context.Context, text string) error {
	if c == nil || c.botToken == "" || c.chatID == "" {
		return nil
	}
	text = truncateForTelegram(text)
	if text == "" {
		return nil
	}
	body := sendMessageRequest{
		ChatID:    c.chatID,
		Text:      text,
		ParseMode: c.parseMode,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"sendMessage", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var result sendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("telegram decode: %w", err)
	}
	if !result.OK {
		return fmt.Errorf("telegram api: %s", result.Description)
	}
	return nil
}
