package main

import (
	"fmt"
	"os"
	"path/filepath"

	"aurumflow/internal/logger"
	"aurumflow/internal/stratrade"
)

func runAuditStrategyTrade(id string) {
	id = filepath.Clean(id)
	if id == "" || id == "." {
		logger.Error("audit-strategy-trade: missing id")
		os.Exit(1)
	}
	root := filepath.Join("research", "strategy-trades")
	a, err := stratrade.Reconstruct(root, id)
	if err != nil {
		logger.Error("audit-strategy-trade: %v", err)
		os.Exit(1)
	}
	fmt.Print(stratrade.FormatAudit(a))
}
