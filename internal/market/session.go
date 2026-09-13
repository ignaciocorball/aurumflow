package market

import (
	"context"
	"encoding/json"
	"fmt"
)

// SessionRequest is the body for POST /api/v1/session.
type SessionRequest struct {
	Identifier       string `json:"identifier"`
	Password         string `json:"password"`
	EncryptedPassword bool   `json:"encryptedPassword,omitempty"`
}

// SessionResponse is the body returned by POST /api/v1/session (accounts etc.).
type SessionResponse struct {
	Accounts         []AccountInfo `json:"accounts"`
	ClientID         string        `json:"clientId"`
	CurrentAccountID string        `json:"currentAccountId"`
	StreamingHost    string        `json:"streamingHost,omitempty"`
}

// AccountInfo holds balance and account id.
type AccountInfo struct {
	AccountID   string  `json:"accountId"`
	AccountName string  `json:"accountName,omitempty"`
	AccountType string  `json:"accountType,omitempty"`
	Currency    string  `json:"currency,omitempty"`
	Preferred   bool    `json:"preferred,omitempty"`
	Status      string  `json:"status,omitempty"`
	Balance     Balance `json:"balance"`
}

// Balance holds balance fields.
type Balance struct {
	Balance   float64 `json:"balance"`
	Available float64 `json:"available"`
	Deposit   float64 `json:"deposit"`
	ProfitLoss float64 `json:"profitLoss"`
}

// CreateSession logs in and sets CST and X-SECURITY-TOKEN on the client.
// Set c.APIKey (from config) before calling.
func (c *Client) CreateSession(ctx context.Context, identifier, password string) (*SessionResponse, error) {
	body := SessionRequest{Identifier: identifier, Password: password}
	var cst, securityToken string
	capture := map[string]*string{"CST": &cst, "X-SECURITY-TOKEN": &securityToken}
	// Session endpoint does not require CST/security token; we need API key from config.
	data, err := c.Do(ctx, "POST", "/api/v1/session", body, capture)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	var sr SessionResponse
	if err := json.Unmarshal(data, &sr); err != nil {
		return nil, fmt.Errorf("parse session response: %w", err)
	}
	c.SetSession(cst, securityToken)
	return &sr, nil
}

// Ping keeps the session alive. Call periodically (e.g. every 5 min).
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.Do(ctx, "GET", "/api/v1/ping", nil, nil)
	return err
}

// SwitchAccountRequest is the body for PUT /api/v1/session (switch active account).
type SwitchAccountRequest struct {
	AccountID string `json:"accountId"`
}

// AccountsResponse is the response from GET /api/v1/accounts.
type AccountsResponse struct {
	Accounts []AccountInfo `json:"accounts"`
}

// GetAccounts returns all accounts (for refresh balance and list). Requires active session.
func (c *Client) GetAccounts(ctx context.Context) (*AccountsResponse, error) {
	data, err := c.Do(ctx, "GET", "/api/v1/accounts", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("get accounts: %w", err)
	}
	var ar AccountsResponse
	if err := json.Unmarshal(data, &ar); err != nil {
		return nil, fmt.Errorf("parse accounts: %w", err)
	}
	return &ar, nil
}

// GetSession returns current session details (GET /api/v1/session). Requires active session.
func (c *Client) GetSession(ctx context.Context) (*SessionResponse, error) {
	data, err := c.Do(ctx, "GET", "/api/v1/session", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	var sr SessionResponse
	if err := json.Unmarshal(data, &sr); err != nil {
		return nil, fmt.Errorf("parse session: %w", err)
	}
	return &sr, nil
}

// SwitchAccount switches the active account and updates session tokens. Call after CreateSession if account_id is set.
func (c *Client) SwitchAccount(ctx context.Context, accountID string) error {
	body := SwitchAccountRequest{AccountID: accountID}
	var cst, securityToken string
	capture := map[string]*string{"CST": &cst, "X-SECURITY-TOKEN": &securityToken}
	_, err := c.Do(ctx, "PUT", "/api/v1/session", body, capture)
	if err != nil {
		return err
	}
	if cst != "" && securityToken != "" {
		c.SetSession(cst, securityToken)
	}
	return nil
}
