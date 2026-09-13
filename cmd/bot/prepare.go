package main

import (
	"context"
	"os"
	"strings"

	"aurumflow/config"
	"aurumflow/internal/logger"
	"aurumflow/internal/money"
)

func runPrepareInstrument(ctx context.Context, epic string, calibrate bool) {
	epic = strings.ToUpper(strings.TrimSpace(epic))
	if epic == "" {
		epic = "GOLD"
	}
	sess, err := bootstrapDemoSession(ctx, config.ExecutionDisabled)
	if err != nil {
		logger.Error("%v", err)
		os.Exit(1)
	}
	details, err := sess.client.GetMarketDetails(ctx, epic)
	if err != nil {
		logger.Error("GetMarketDetails %s: %v", epic, err)
		os.Exit(1)
	}
	spec := money.FromMarketDetails(details)
	logger.Info("PREPARE epic=%s name=%s type=%s currency=%s status=%s lot=%.6f contract=%.6f scale=%.6f pip=%d tick=%.6f min=%.6f max=%.6f inc=%.6f mpu=%.6f validation=%s",
		spec.Epic, emptyDash(spec.Name), emptyDash(spec.InstrumentType), emptyDash(spec.Currency),
		details.Snapshot.MarketStatus, spec.LotSize, spec.ContractSize, spec.ScalingFactor, spec.PipPosition, spec.TickSize,
		spec.MinDealSize, spec.MaxDealSize, spec.SizeIncrement, spec.MoneyPerPriceUnit, spec.ValidationStatus)
	cachePath := money.CachePath("", epic)
	if prev, err := money.LoadSpec(cachePath); err == nil {
		if reason := money.InvalidateReason(prev, spec); reason != "" {
			logger.Warn("CACHE INVALID: %s", reason)
		} else if prev.ValidationStatus == money.RuntimeValidated {
			spec.ValidationStatus = prev.ValidationStatus
			spec.ValidationEvidence = prev.ValidationEvidence
			spec.ValidatedAt = prev.ValidatedAt
			spec.MoneyPerPriceUnit = prev.MoneyPerPriceUnit
			logger.Info("CACHE: reused RUNTIME_VALIDATED spec for %s", epic)
		}
	}
	if !marketAllowsExecution(details.Snapshot.MarketStatus) {
		_ = money.SaveSpec(cachePath, spec)
		logger.Info("PREPARE: market %s is %s — persisted BROKER_METADATA_ONLY; no mutation", epic, details.Snapshot.MarketStatus)
		return
	}
	if !spec.DealingComplete() {
		logger.Error("PREPARE: dealing rules incomplete")
		os.Exit(1)
	}
	if calibrate || spec.ValidationStatus != money.RuntimeValidated {
		if !demoEnvConfigured() {
			logger.Error("calibration requires DEMO credentials")
			os.Exit(1)
		}
		logger.Info("PREPARE: running controlled min-size calibration canary for %s", epic)
		runLifecycleCanary(ctx, epic)
		if fresh, err := sess.client.GetPositions(ctx); err == nil {
			left := 0
			for _, p := range fresh.Positions {
				if strings.EqualFold(p.GetEpic(), epic) {
					left++
				}
			}
			if left != 0 {
				logger.Error("PREPARE HALT: %s still has %d open positions", epic, left)
				os.Exit(1)
			}
		}
		// Single-trade close PnL is insufficient. Keep metadata inference unless UPL samples exist.
		if spec.MoneyPerPriceUnit > 0 {
			spec.ValidationEvidence = "broker metadata inferred; calibration canary closed to zero; UPL multi-sample still required for RUNTIME_VALIDATED"
		}
	}
	_ = money.SaveSpec(cachePath, spec)
	logger.Info("PREPARE DONE epic=%s validation=%s cache=%s", spec.Epic, spec.ValidationStatus, cachePath)
}
