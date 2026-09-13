package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"aurumflow/config"
	"aurumflow/internal/logger"
	"aurumflow/internal/market"
	"aurumflow/internal/risk"
	"aurumflow/pkg/models"
)

func demoEnvConfigured() bool {
	return os.Getenv("AURUMFLOW_DEMO_API_KEY") != "" &&
		os.Getenv("AURUMFLOW_DEMO_IDENTIFIER") != "" &&
		os.Getenv("AURUMFLOW_DEMO_PASSWORD") != ""
}

func loadDemoCanaryConfig(execMode config.ExecutionMode) (*config.Config, error) {
	path := os.Getenv("AURUMFLOW_CONFIG")
	if strings.TrimSpace(path) == "" {
		path = "config/demo_config.json"
	}
	cfg := &config.Config{}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if config.IsLiveCapitalHost(cfg.API.BaseURL) || strings.EqualFold(cfg.API.Environment, "live") || strings.EqualFold(cfg.API.Mode, "live") {
		return nil, fmt.Errorf("FATAL: LIVE broker access is disabled in P1. demo_config must not declare a LIVE host or environment")
	}
	// Credentials and LIVE account_id never come from JSON.
	cfg.API.APIKey = os.Getenv("AURUMFLOW_DEMO_API_KEY")
	cfg.API.Identifier = os.Getenv("AURUMFLOW_DEMO_IDENTIFIER")
	cfg.API.Password = os.Getenv("AURUMFLOW_DEMO_PASSWORD")
	cfg.API.AccountID = strings.TrimSpace(os.Getenv("AURUMFLOW_DEMO_ACCOUNT_ID"))
	cfg.API.Environment = "demo"
	cfg.API.Mode = "demo"
	if execMode == "" {
		execMode = config.ExecutionDisabled
	}
	cfg.API.ExecutionMode = string(execMode)
	cfg.API.BaseURL = ""
	if cfg.Notifications != nil {
		cfg.Notifications.Enabled = false
	}
	if cfg.Telemetry != nil && cfg.Telemetry.Firebase != nil {
		cfg.Telemetry.Firebase.Enabled = false
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func runReadOnlyCanary(ctx context.Context) {
	if !demoEnvConfigured() {
		logger.Error("DEMO_RUNTIME = BLOCKED_MISSING_DEMO_CREDENTIALS")
		logger.Error("Need AURUMFLOW_DEMO_API_KEY, AURUMFLOW_DEMO_IDENTIFIER, AURUMFLOW_DEMO_PASSWORD (do not reuse LIVE secrets from JSON files)")
		os.Exit(2)
	}
	cfg, err := loadDemoCanaryConfig(config.ExecutionDisabled)
	if err != nil {
		logger.Error("%v", err)
		os.Exit(1)
	}
	client := market.NewDemoClient()
	client.APIKey = cfg.API.APIKey
	if config.IsLiveCapitalHost(client.BaseURL) || !config.IsDemoCapitalHost(client.BaseURL) {
		logger.Error("FATAL: LIVE broker access is disabled in P1.")
		os.Exit(1)
	}
	if cfg.API.BaseURL != config.DemoAPIHost || !config.IsDemoCapitalHost(cfg.API.BaseURL) {
		logger.Error("FATAL: LIVE broker access is disabled in P1.")
		os.Exit(1)
	}
	if cfg.ExecMode() != config.ExecutionDisabled {
		logger.Error("canary-readonly requires EXECUTION MODE DISABLED (got %s)", cfg.ExecMode())
		os.Exit(1)
	}

	logger.Info("API ENVIRONMENT: %s", cfg.API.Environment)
	logger.Info("API BASE URL: %s", client.BaseURL)
	logger.Info("EXECUTION MODE: %s", cfg.ExecMode())
	logger.Info("live requests = impossible (host hard-gate + DISABLED)")

	session, err := client.CreateSession(ctx, cfg.API.Identifier, cfg.API.Password)
	if err != nil {
		logger.Error("AUTH: FAIL (%v)", err)
		os.Exit(1)
	}
	if config.IsLiveCapitalHost(client.BaseURL) {
		logger.Error("FATAL: LIVE broker access is disabled in P1.")
		os.Exit(1)
	}
	logger.Info("AUTH: PASS")

	if err := client.Ping(ctx); err != nil {
		logger.Error("PING: FAIL (%v)", err)
		os.Exit(1)
	}
	logger.Info("PING: PASS")

	ar, err := client.GetAccounts(ctx)
	if err != nil {
		logger.Error("ACCOUNT: FAIL (%v)", err)
		os.Exit(1)
	}
	logger.Info("DEMO ACCOUNTS FOUND: %d", len(ar.Accounts))
	for i, a := range ar.Accounts {
		logger.Info("DEMO ACCOUNT[%d] id=%s name=%s type=%s currency=%s preferred=%v status=%s balance=%.2f",
			i, config.RedactAccountID(a.AccountID), emptyDash(a.AccountName), emptyDash(a.AccountType),
			emptyDash(a.Currency), a.Preferred, emptyDash(a.Status), a.Balance.Balance)
	}

	acc, err := risk.SelectDemoAccount(ar.Accounts, session.CurrentAccountID)
	if err != nil {
		logger.Error("ACCOUNT: FAIL (%v)", err)
		os.Exit(1)
	}
	if err := os.Setenv("AURUMFLOW_DEMO_ACCOUNT_ID", acc.AccountID); err != nil {
		logger.Error("ACCOUNT: FAIL set process account id: %v", err)
		os.Exit(1)
	}
	if acc.AccountID != session.CurrentAccountID {
		if err := client.SwitchAccount(ctx, acc.AccountID); err != nil {
			logger.Error("ACCOUNT: FAIL switch (%v)", err)
			os.Exit(1)
		}
		ar, err = client.GetAccounts(ctx)
		if err != nil {
			logger.Error("ACCOUNT: FAIL refresh (%v)", err)
			os.Exit(1)
		}
		acc, err = risk.SelectAccount(ar.Accounts, acc.AccountID, acc.AccountID)
		if err != nil {
			logger.Error("ACCOUNT: FAIL (%v)", err)
			os.Exit(1)
		}
	}
	logger.Info("ACCOUNT: PASS id=%s name=%s type=%s currency=%s preferred=%v status=%s balance=%.2f",
		config.RedactAccountID(acc.AccountID), emptyDash(acc.AccountName), emptyDash(acc.AccountType),
		emptyDash(acc.Currency), acc.Preferred, emptyDash(acc.Status), acc.Balance.Balance)

	epic := os.Getenv("AURUMFLOW_EPIC")
	if epic == "" {
		epic, err = client.ResolveEpic(ctx, "gold")
		if err != nil {
			logger.Error("MARKET DETAILS: FAIL resolve (%v)", err)
			os.Exit(1)
		}
	}
	details, err := client.GetMarketDetails(ctx, epic)
	if err != nil {
		logger.Error("MARKET DETAILS: FAIL (%v)", err)
		os.Exit(1)
	}
	spec, specErr := market.SpecFromDetails(details, cfg.Risk.ValuePerPoint)
	logger.Info("MARKET DETAILS: PASS")
	logger.Info("GOLD provider=Capital.com epic=%s name=%s type=%s currency=%s status=%s min=%.6f max=%.6f increment=%.6f lot=%.6f spec_complete=%v spec_err=%v",
		spec.Epic, emptyDash(spec.Name), emptyDash(spec.InstrumentType), emptyDash(spec.Currency),
		emptyDash(spec.MarketStatus), spec.MinDealSize, spec.MaxDealSize, spec.SizeStep, spec.LotSize,
		spec.SizingComplete(), specErr)
	logger.Info("GOLD identity: Capital GOLD CFD != CME GC futures")

	now := time.Now().UTC()
	candles, err := client.GetPrices(ctx, epic, market.ResolutionMinute15, 20, now.Add(-24*time.Hour), now)
	if err != nil {
		logger.Info("PRICES: 24h window empty or closed (%v); retrying last available M15", err)
		candles, err = client.GetPrices(ctx, epic, market.ResolutionMinute15, 20, time.Time{}, time.Time{})
	}
	if err != nil {
		logger.Info("PRICES: unbounded M15 failed (%v); retrying 7d H1", err)
		candles, err = client.GetPrices(ctx, epic, market.ResolutionHour, 20, now.Add(-7*24*time.Hour), now)
	}
	if err != nil {
		logger.Error("PRICES: FAIL (%v)", err)
		os.Exit(1)
	}
	logger.Info("PRICES: PASS candles=%d last=%s", len(candles), formatCanaryCandle(candles))

	pos, err := client.GetPositions(ctx)
	if err != nil {
		logger.Error("POSITIONS: FAIL (%v)", err)
		os.Exit(1)
	}
	openN := 0
	if pos != nil {
		openN = len(pos.Positions)
	}
	logger.Info("POSITIONS: PASS open=%d", openN)

	if err := client.Ping(ctx); err != nil {
		logger.Error("PING: FAIL (%v)", err)
		os.Exit(1)
	}
	logger.Info("PING: PASS")
	logger.Info("MUTATING REQUESTS: 0 (no POST/PUT/DELETE /positions)")
	logger.Info("DEMO read-only canary complete (no orders sent)")
}

func discoverTradeableEpic(ctx context.Context, client *market.Client, prefer string) (string, string, error) {
	prefer = strings.TrimSpace(prefer)
	if prefer != "" {
		return prefer, "explicit", nil
	}
	for _, term := range []string{"Bitcoin", "Ethereum"} {
		ms, err := client.SearchMarkets(ctx, term)
		if err != nil {
			logger.Info("search %s: %v", term, err)
			continue
		}
		logger.Info("search %s: %d markets", term, len(ms))
		m, err := market.PickTradeableCanary(ms)
		if err != nil {
			logger.Info("search %s: no TRADEABLE (%v)", term, err)
			continue
		}
		return m.Epic, "search:"+term, nil
	}
	return "", "", fmt.Errorf("no TRADEABLE Bitcoin/Ethereum instrument on this DEMO account")
}

func marketAllowsExecution(status string) bool {
	s := strings.ToUpper(strings.TrimSpace(status))
	return s == "TRADEABLE"
}

func canaryDealSize(min, increment, max float64) (float64, error) {
	if min <= 0 || increment <= 0 || max < min {
		return 0, fmt.Errorf("dealing rules incomplete or ambiguous: min=%.6f increment=%.6f max=%.6f", min, increment, max)
	}
	size := min
	n := size / increment
	aligned := float64(int(n+1e-9)) * increment
	if aligned+1e-12 < min {
		aligned += increment
	}
	if aligned+1e-12 < min || aligned > max+1e-12 {
		return 0, fmt.Errorf("cannot form a valid canary size from min=%.6f increment=%.6f max=%.6f", min, increment, max)
	}
	return aligned, nil
}

func formatCanaryCandle(candles []models.Candle) string {
	if len(candles) == 0 {
		return "none"
	}
	return candles[len(candles)-1].Time.UTC().Format(time.RFC3339)
}

func emptyDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}
