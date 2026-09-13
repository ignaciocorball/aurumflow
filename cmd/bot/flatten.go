package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"aurumflow/config"
	"aurumflow/internal/execution"
	"aurumflow/internal/journal"
	"aurumflow/internal/logger"
	"aurumflow/internal/market"
)

func runFlatten(ctx context.Context, cfg *config.Config, epic string, confirmed bool) {
	if cfg.ExecMode() != config.ExecutionDemo {
		logger.Error("flatten refused: execution_mode must be DEMO (got %s)", cfg.ExecMode())
		os.Exit(1)
	}
	if config.IsLiveCapitalHost(cfg.API.BaseURL) || !config.IsDemoCapitalHost(cfg.API.BaseURL) {
		logger.Error("FATAL: LIVE broker access is disabled in P1.")
		os.Exit(1)
	}
	if !confirmed {
		fmt.Fprintf(os.Stderr, "This will close ALL open DEMO positions for epic %s.\nType FLATTEN-DEMO to confirm: ", epic)
		line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		if strings.TrimSpace(line) != "FLATTEN-DEMO" {
			logger.Error("flatten aborted")
			os.Exit(1)
		}
	}

	client := market.NewDemoClient()
	client.APIKey = cfg.API.APIKey
	if _, err := client.CreateSession(ctx, cfg.API.Identifier, cfg.API.Password); err != nil {
		logger.Error("flatten session: %v", err)
		os.Exit(1)
	}
	if cfg.API.AccountID != "" {
		if err := client.SwitchAccount(ctx, cfg.API.AccountID); err != nil {
			logger.Error("flatten switch account: %v", err)
			os.Exit(1)
		}
	}

	ex := execution.NewExecutor(client, epic, 0.1, 0.1)
	positions, err := ex.Positions(ctx)
	if err != nil {
		logger.Error("flatten list: %v", err)
		os.Exit(1)
	}
	n := 0
	for _, p := range positions {
		if epic != "" && p.Epic != "" && p.Epic != epic {
			continue
		}
		logger.Info("flatten: closing dealId=%s epic=%s", p.DealID, p.Epic)
		if _, err := ex.ClosePosition(ctx, p.DealID); err != nil {
			logger.Error("flatten close %s: %v", p.DealID, err)
			os.Exit(1)
		}
		n++
	}
	logger.Info("flatten complete: closed=%d epic=%s event=%s", n, epic, journal.EventPositionClosed)
}
