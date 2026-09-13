package main

import (
	"context"
	"fmt"
	"os"

	"aurumflow/config"
	"aurumflow/internal/logger"
	"aurumflow/internal/market"
	"aurumflow/internal/risk"
)

type demoSession struct {
	cfg    *config.Config
	client *market.Client
	acc    market.AccountInfo
}

func bootstrapDemoSession(ctx context.Context, execMode config.ExecutionMode) (*demoSession, error) {
	if !demoEnvConfigured() {
		return nil, fmt.Errorf("DEMO_RUNTIME = BLOCKED_MISSING_DEMO_CREDENTIALS")
	}
	cfg, err := loadDemoCanaryConfig(execMode)
	if err != nil {
		return nil, err
	}
	if config.IsLiveCapitalHost(cfg.API.BaseURL) || !config.IsDemoCapitalHost(cfg.API.BaseURL) {
		return nil, fmt.Errorf("FATAL: LIVE broker access is disabled")
	}
	client := market.NewDemoClient()
	client.APIKey = cfg.API.APIKey
	if config.IsLiveCapitalHost(client.BaseURL) || !config.IsDemoCapitalHost(client.BaseURL) {
		return nil, fmt.Errorf("FATAL: LIVE broker access is disabled")
	}
	session, err := client.CreateSession(ctx, cfg.API.Identifier, cfg.API.Password)
	if err != nil {
		return nil, fmt.Errorf("AUTH FAIL: %w", err)
	}
	ar, err := client.GetAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("ACCOUNT FAIL: %w", err)
	}
	acc, err := risk.SelectDemoAccount(ar.Accounts, session.CurrentAccountID)
	if err != nil {
		return nil, err
	}
	_ = os.Setenv("AURUMFLOW_DEMO_ACCOUNT_ID", acc.AccountID)
	if acc.AccountID != session.CurrentAccountID {
		if err := client.SwitchAccount(ctx, acc.AccountID); err != nil {
			return nil, fmt.Errorf("ACCOUNT switch FAIL: %w", err)
		}
	}
	logger.Info("AUTH: PASS account=%s type=%s currency=%s balance=%.2f",
		config.RedactAccountID(acc.AccountID), emptyDash(acc.AccountType), emptyDash(acc.Currency), acc.Balance.Balance)
	return &demoSession{cfg: cfg, client: client, acc: acc}, nil
}
