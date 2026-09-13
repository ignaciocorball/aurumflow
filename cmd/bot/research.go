package main

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"

	"aurumflow/internal/binancehist"
	"aurumflow/internal/logger"
	"aurumflow/internal/research"
)

func runResearchBTC(ctx context.Context, fromS, toS string, days int) {
	if days <= 0 {
		days = 30
	}
	to := time.Now().UTC().AddDate(0, 0, -1)
	from := to.AddDate(0, 0, -(days - 1))
	if fromS != "" {
		if t, err := time.Parse("2006-01-02", fromS); err == nil {
			from = t
		}
	}
	if toS != "" {
		if t, err := time.Parse("2006-01-02", toS); err == nil {
			to = t
		}
	}
	arch := binancehist.New("data", "BTCUSDT")
	var paths []string
	for _, day := range binancehist.DaysInclusive(from, to) {
		p := arch.DayPath(day)
		if _, err := os.Stat(p); err != nil {
			got, err := arch.DownloadDay(ctx, day)
			if err != nil {
				logger.Warn("missing %s: %v", day, err)
				continue
			}
			p = got
		}
		paths = append(paths, p)
	}
	if len(paths) == 0 {
		logger.Error("no historical days available")
		os.Exit(1)
	}
	logger.Info("research btc-radar days=%d first=%s last=%s", len(paths), paths[0], paths[len(paths)-1])
	res, err := research.RunFromZips(ctx, paths, "BTCUSDT")
	if err != nil {
		logger.Error("research: %v", err)
		os.Exit(1)
	}
	commit := gitHead()
	if err := research.WriteReports("research/reports", res, commit); err != nil {
		logger.Error("report: %v", err)
		os.Exit(1)
	}
	logger.Info("research done events=%d legacy=%d aligned=%d contradicted=%d runtime=%s report=research/reports/FREE_RESEARCH_BASELINE.md",
		res.TradeEvents, res.LegacySignals, res.Aligned, res.Contradicted, res.Runtime)
}

func gitHead() string {
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}
