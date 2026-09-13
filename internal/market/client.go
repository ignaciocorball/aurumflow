package market

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"aurumflow/config"

	"golang.org/x/time/rate"
)

// Default rate limit: 5 req/s, burst 2 (Capital.com allows max 10 req/s per user).
const (
	defaultRateLimit = 5
	defaultBurst    = 2
)

// Client is the Capital.com REST API client.
type Client struct {
	BaseURL      string
	HTTPClient   *http.Client
	CST          string
	SecurityToken string
	APIKey       string
	limiter      *rate.Limiter
}

// NewClient creates a new API client with default rate limiting (5 req/s, burst 2).
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		limiter: rate.NewLimiter(rate.Limit(defaultRateLimit), defaultBurst),
	}
}

// NewDemoClient returns a client pinned to the Capital.com DEMO host.
func NewDemoClient() *Client {
	return NewClient(config.DemoAPIHost)
}

// SetSession sets the session tokens from a successful login.
func (c *Client) SetSession(cst, securityToken string) {
	c.CST = cst
	c.SecurityToken = securityToken
}

// Do sends an authenticated request. If no session, only X-CAP-API-KEY is sent (for session create).
// Requests are rate-limited (default 5 req/s, burst 2). On 429 Too Many Requests, Do retries with backoff.
func (c *Client) Do(ctx context.Context, method, path string, body interface{}, captureHeaders map[string]*string) ([]byte, error) {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
	}

	const maxRetries = 3
	backoffDurations := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second}
	var lastData []byte

	if err := c.assertNotLive(path); err != nil {
		return nil, err
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := c.limiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limit wait: %w", err)
		}

		var bodyReader io.Reader
		if len(bodyBytes) > 0 {
			bodyReader = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("new request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		if c.APIKey != "" {
			req.Header.Set("X-CAP-API-KEY", c.APIKey)
		}
		if c.CST != "" {
			req.Header.Set("CST", c.CST)
		}
		if c.SecurityToken != "" {
			req.Header.Set("X-SECURITY-TOKEN", c.SecurityToken)
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("do request: %w", err)
		}
		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read body: %w", readErr)
		}

		if resp.StatusCode == 401 {
			return nil, fmt.Errorf("api 401 unauthorized: session expired or invalid; re-login required")
		}
		if resp.StatusCode != 429 {
			if captureHeaders != nil {
				if v := resp.Header.Get("CST"); v != "" && captureHeaders["CST"] != nil {
					*captureHeaders["CST"] = v
				}
				if v := resp.Header.Get("X-SECURITY-TOKEN"); v != "" && captureHeaders["X-SECURITY-TOKEN"] != nil {
					*captureHeaders["X-SECURITY-TOKEN"] = v
				}
			}
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, string(data))
			}
			return data, nil
		}

		lastData = data
		if attempt == maxRetries {
			break
		}
		delay := backoffDurations[attempt]
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			if sec, err := strconv.Atoi(retryAfter); err == nil && sec > 0 && sec <= 60 {
				delay = time.Duration(sec) * time.Second
			}
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}

	return nil, fmt.Errorf("api 429 too many requests after %d retries: %s", maxRetries+1, string(lastData))
}
