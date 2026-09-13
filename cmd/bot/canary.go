package main

import (
	"context"
	"os"
	"time"

	"aurumflow/config"
	"aurumflow/internal/logger"
	"aurumflow/internal/market"
	"aurumflow/internal/risk"
)

func demoEnvConfigured() bool {
	return os.Getenv("AURUMFLOW_DEMO_API_KEY") != "" &&
		os.Getenv("AURUMFLOW_DEMO_IDENTIFIER") != "" &&
		os.Getenv("AURUMFLOW_DEMO_PASSWORD") != ""
}

func runReadOnlyCanary(ctx context.Context) {
	if !demoEnvConfigured() {
		logger.Error("DEMO_RUNTIME = BLOCKED_MISSING_DEMO_CREDENTIALS")
		logger.Error("Need AURUMFLOW_DEMO_API_KEY, AURUMFLOW_DEMO_IDENTIFIER, AURUMFLOW_DEMO_PASSWORD (do not reuse LIVE secrets)")
		os.Exit(2)
	}
	cfg := &config.Config{API: config.APIConfig{
		Environment:   "demo",
		Mode:          "demo",
		ExecutionMode: "DISABLED",
		APIKey:        os.Getenv("AURUMFLOW_DEMO_API_KEY"),
		Identifier:    os.Getenv("AURUMFLOW_DEMO_IDENTIFIER"),
		Password:      os.Getenv("AURUMFLOW_DEMO_PASSWORD"),
		AccountID:     os.Getenv("AURUMFLOW_DEMO_ACCOUNT_ID"),
	}}
	if err := cfg.Validate(); err != nil {
		logger.Error("%v", err)
		os.Exit(1)
	}
	client := market.NewDemoClient()
	client.APIKey = cfg.API.APIKey
	if config.IsLiveCapitalHost(client.BaseURL) {
		logger.Error("FATAL: LIVE broker access is disabled in P1.")
		os.Exit(1)
	}

	session, err := client.CreateSession(ctx, cfg.API.Identifier, cfg.API.Password)
	if err != nil {
		logger.Error("DEMO AUTH: FAIL (%v)", err)
		os.Exit(1)
	}
	logger.Info("DEMO AUTH: PASS")

	ar, err := client.GetAccounts(ctx)
	if err != nil {
		logger.Error("ACCOUNT: FAIL (%v)", err)
		os.Exit(1)
	}
	if cfg.API.AccountID != "" {
		if err := client.SwitchAccount(ctx, cfg.API.AccountID); err != nil {
			logger.Error("ACCOUNT: FAIL switch (%v)", err)
			os.Exit(1)
		}
		ar, err = client.GetAccounts(ctx)
		if err != nil {
			logger.Error("ACCOUNT: FAIL refresh (%v)", err)
			os.Exit(1)
		}
	}
	acc, err := risk.SelectAccount(ar.Accounts, cfg.API.AccountID, session.CurrentAccountID)
	if err != nil {
		logger.Error("ACCOUNT: FAIL (%v)", err)
		os.Exit(1)
	}
	logger.Info("ACCOUNT: PASS id=%s", config.RedactAccountID(acc.AccountID))

	epic := os.Getenv("AURUMFLOW_EPIC")
	if epic == "" {
		epic, err = client.ResolveEpic(ctx, "gold")
		if err != nil {
			logger.Error("MARKET DATA: FAIL resolve (%v)", err)
			os.Exit(1)
		}
	}
	details, err := client.GetMarketDetails(ctx, epic)
	if err != nil {
		logger.Error("MARKET DETAILS: FAIL (%v)", err)
		os.Exit(1)
	}
	spec, err := market.SpecFromDetails(details, cfg.Risk.ValuePerPoint)
	logger.Info("MARKET DETAILS: PASS epic=%s complete=%v spec_err=%v", epic, spec.SizingComplete(), err)

	now := time.Now().UTC()
	if _, err := client.GetPrices(ctx, epic, market.ResolutionMinute15, 20, now.Add(-24*time.Hour), now); err != nil {
		logger.Error("PRICES: FAIL (%v)", err)
		os.Exit(1)
	}
	logger.Info("MARKET DATA: PASS")

	if _, err := client.GetPositions(ctx); err != nil {
		logger.Error("POSITIONS READ: FAIL (%v)", err)
		os.Exit(1)
	}
	logger.Info("POSITIONS READ: PASS")

	if err := client.Ping(ctx); err != nil {
		logger.Error("PING: FAIL (%v)", err)
		os.Exit(1)
	}
	logger.Info("PING: PASS")
	logger.Info("DEMO read-only canary complete (no orders sent)")
}
