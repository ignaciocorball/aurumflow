package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"aurumflow/config"
	"aurumflow/internal/execacct"
	"aurumflow/internal/logger"
	"aurumflow/internal/market"
)

type demoSession struct {
	cfg      *config.Config
	client   *market.Client
	acc      market.AccountInfo
	Identity execacct.Identity
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
	explicit := strings.TrimSpace(os.Getenv("AURUMFLOW_DEMO_ACCOUNT_ID"))
	if explicit == "" {
		explicit = strings.TrimSpace(cfg.API.AccountID)
	}
	ident := execacct.Resolve(ar.Accounts, explicit, "")
	var acc market.AccountInfo
	if ident.AccountID != "" {
		for _, a := range ar.Accounts {
			if a.AccountID == ident.AccountID {
				acc = a
				break
			}
		}
	}
	if ident.MayTrade {
		cfg.API.AccountID = ident.AccountID
		if acc.AccountID != "" && acc.AccountID != session.CurrentAccountID {
			if err := client.SwitchAccount(ctx, acc.AccountID); err != nil {
				return nil, fmt.Errorf("ACCOUNT switch FAIL: %w", err)
			}
		}
		logger.Info("AUTH: PASS account=%s type=%s currency=%s status=VERIFIED",
			execacct.Mask(ident.AccountID), emptyDash(ident.Type), emptyDash(ident.Currency))
	} else {
		logger.Warn("ACCOUNT %s: %s", ident.Resolution, execacct.Instruction(ident))
		if session.CurrentAccountID != "" {
			for _, a := range ar.Accounts {
				if a.AccountID == session.CurrentAccountID {
					acc = a
					break
				}
			}
		}
	}
	return &demoSession{cfg: cfg, client: client, acc: acc, Identity: ident}, nil
}
