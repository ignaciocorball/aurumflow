package notifications

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"aurumflow/config"
)

const (
	defaultPushoverURL = "https://api.pushover.net/1/messages.json"
	maxRetries         = 3
	retryBackoff       = 2 * time.Second
)

// PushoverClient sends messages to the Pushover API.
type PushoverClient struct {
	apiURL                 string
	token                  string
	user                   string
	device                 string
	sound                  string
	defaultPriority        int
	emergencyRetrySeconds  int
	emergencyExpireSeconds int
	client                 *http.Client
}

// NewPushoverClient builds a client from config. Token and user must be non-empty for sending.
func NewPushoverClient(cfg *config.PushoverConfig) *PushoverClient {
	if cfg == nil {
		return nil
	}
	apiURL := cfg.APIURL
	if apiURL == "" {
		apiURL = defaultPushoverURL
	}
	return &PushoverClient{
		apiURL:                 apiURL,
		token:                  cfg.Token,
		user:                   cfg.User,
		device:                 cfg.Device,
		sound:                  cfg.Sound,
		defaultPriority:        cfg.DefaultPriority,
		emergencyRetrySeconds:  cfg.EmergencyRetrySeconds,
		emergencyExpireSeconds: cfg.EmergencyExpireSeconds,
		client:                 &http.Client{Timeout: 15 * time.Second},
	}
}

// Send posts a message to Pushover. Handles 429 (retry-after) and 5xx with limited retries.
// For priority=2, retry and expire are sent as per config.
func (c *PushoverClient) Send(ctx context.Context, title, message string, priority int) error {
	if c == nil || c.token == "" || c.user == "" {
		return nil
	}
	form := url.Values{}
	form.Set("token", c.token)
	form.Set("user", c.user)
	form.Set("message", message)
	if title != "" {
		form.Set("title", title)
	}
	if c.device != "" {
		form.Set("device", c.device)
	}
	if c.sound != "" {
		form.Set("sound", c.sound)
	}
	form.Set("priority", strconv.Itoa(priority))
	if priority == 2 {
		if c.emergencyRetrySeconds > 0 {
			form.Set("retry", strconv.Itoa(c.emergencyRetrySeconds))
		}
		if c.emergencyExpireSeconds > 0 {
			form.Set("expire", strconv.Itoa(c.emergencyExpireSeconds))
		}
	}
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryBackoff):
			}
		}
			body := form.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL, strings.NewReader(body))
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.ContentLength = int64(len(body))

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		_ = resp.Body.Close()
		switch resp.StatusCode {
		case http.StatusOK:
			return nil
		case http.StatusTooManyRequests:
			if s := resp.Header.Get("Retry-After"); s != "" {
				if sec, _ := strconv.Atoi(s); sec > 0 && sec < 120 {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(time.Duration(sec) * time.Second):
					}
				}
			}
			lastErr = fmt.Errorf("pushover 429 rate limit")
			continue
		default:
			if resp.StatusCode >= 500 {
				lastErr = fmt.Errorf("pushover %d", resp.StatusCode)
				continue
			}
			lastErr = fmt.Errorf("pushover %d", resp.StatusCode)
			return lastErr
		}
	}
	return lastErr
}
