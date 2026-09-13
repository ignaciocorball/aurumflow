package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aurumflow/internal/logger"
)

// runSundayRuntime supervises SHADOW intelligence + console only.
// It does not calibrate GOLD, start demo-week, or bypass execution gates.
func runSundayRuntime(ctx context.Context, soakMin int, l2 string) {
	if soakMin <= 0 {
		soakMin = 24 * 60
	}
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		cancel()
	}()
	logger.Info("sunday-runtime: shadow intelligence on 127.0.0.1:8766 — GOLD execution is NOT started")
	logger.Info("prepare GOLD and --demo-week remain explicit operator steps")
	runShadowRuntime(runCtx, time.Duration(soakMin)*time.Minute, "127.0.0.1:8766", l2)
}
