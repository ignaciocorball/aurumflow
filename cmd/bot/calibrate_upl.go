package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aurumflow/config"
	"aurumflow/internal/execacct"
	"aurumflow/internal/execution"
	"aurumflow/internal/journal"
	"aurumflow/internal/killswitch"
	"aurumflow/internal/logger"
	"aurumflow/internal/market"
	"aurumflow/internal/money"
)

var (
	calibMethod  = "upl-slope"
	calibSeconds = 150
)

func runUPLSlopeCalibration(ctx context.Context, epic string, spec money.MonetaryInstrumentSpec) (money.MonetaryInstrumentSpec, money.CalibrationResult) {
	var last money.CalibrationResult
	attempts := 0
	for attempts < money.MaxCalibCanaries {
		attempts++
		res := runOneUPLCanary(ctx, epic, spec, attempts)
		last = res
		if res.OK {
			return money.ApplyRuntime(spec, res), res
		}
		logger.Info("CALIB attempt=%d result=%s mpu=%.6f samples=%d pxRange=%.6f uplRange=%.6f",
			attempts, res.Failure, res.MoneyPerPriceUnit, res.Samples, res.PriceRange, res.UPLRange)
		if res.Failure != money.FailInsufficientPriceExcursion {
			break
		}
		logger.Info("CALIB retrying once after INSUFFICIENT_PRICE_EXCURSION")
	}
	return money.ApplyRuntime(spec, last), last
}

func runOneUPLCanary(ctx context.Context, epic string, spec money.MonetaryInstrumentSpec, attempt int) (res money.CalibrationResult) {
	res.Failure = money.FailBrokerUncertain
	if !demoEnvConfigured() {
		logger.Error("calibration requires DEMO credentials")
		return res
	}
	cfg, err := loadDemoCanaryConfig(config.ExecutionDemo)
	if err != nil {
		logger.Error("%v", err)
		return res
	}
	ks := killswitch.New(false, killswitch.DefaultFile)
	if ks.HaltNewOrders() {
		logger.Error("PRECHECK FAIL: kill switch ON")
		return res
	}
	client := market.NewDemoClient()
	client.APIKey = cfg.API.APIKey
	if config.IsLiveCapitalHost(client.BaseURL) || !config.IsDemoCapitalHost(client.BaseURL) {
		logger.Error("FATAL: LIVE broker access is disabled")
		return res
	}
	session, err := client.CreateSession(ctx, cfg.API.Identifier, cfg.API.Password)
	if err != nil {
		logger.Error("AUTH: FAIL (%v)", err)
		return res
	}
	if ar, aerr := client.GetAccounts(ctx); aerr == nil {
		ident := execacct.Resolve(ar.Accounts, strings.TrimSpace(os.Getenv("AURUMFLOW_DEMO_ACCOUNT_ID")), "")
		if ident.MayTrade && ident.AccountID != session.CurrentAccountID {
			_ = client.SwitchAccount(ctx, ident.AccountID)
		}
	}
	details, err := client.GetMarketDetails(ctx, epic)
	if err != nil || details == nil {
		logger.Error("MARKET DETAILS: FAIL (%v)", err)
		return res
	}
	if !marketAllowsExecution(details.Snapshot.MarketStatus) {
		logger.Error("CALIB BLOCKED_MARKET_CLOSED status=%s", details.Snapshot.MarketStatus)
		return res
	}
	size, err := canaryDealSize(details.DealingRules.MinDealSize.Value, details.DealingRules.MinSizeIncrement.Value, details.DealingRules.MaxDealSize.Value)
	if err != nil {
		logger.Error("SIZE: FAIL (%v)", err)
		return res
	}
	pos, err := client.GetPositions(ctx)
	if err != nil {
		logger.Error("POSITIONS: FAIL (%v)", err)
		return res
	}
	for _, p := range pos.Positions {
		if strings.EqualFold(p.GetEpic(), epic) {
			logger.Error("PRECHECK FAIL: existing %s positions", epic)
			res.Failure = money.FailPositionReconcile
			return res
		}
	}
	jw, err := journal.NewFileWriter(filepath.Join("journals", "canary-lifecycle.jsonl"))
	if err != nil {
		logger.Error("journal: %v", err)
		return res
	}
	defer func() { _ = jw.Close() }()

	ex := execution.NewExecutor(client, epic, details.DealingRules.MinDealSize.Value, details.DealingRules.MinSizeIncrement.Value)
	openRes, err := ex.OpenInfrastructure(ctx, canaryDir, size)
	if err != nil {
		logger.Error("OPEN: FAIL (%v) — not retrying POST", err)
		return res
	}
	conf, err := pollConfirm(ctx, ex, openRes.DealReference)
	if err != nil || rejectedConfirm(conf) {
		logger.Error("CONFIRM: FAIL (%v) — not opening another", err)
		reportGoldPositions(ctx, client, epic)
		return res
	}
	item, err := findPosition(ctx, client, conf.DealID, epic)
	if err != nil {
		logger.Error("READ: FAIL (%v)", err)
		if conf.DealID != "" {
			_, _ = closeCanary(ctx, ex, ks, jw, epic, conf.DealID, size, conf.Direction)
		}
		return res
	}
	dealID := item.Position.DealID
	dir := item.Position.Direction
	if item.Position.Size > 0 {
		size = item.Position.Size
	}
	logger.Info("CALIB OPEN attempt=%d dealId=%s dir=%s size=%.4f level=%.2f", attempt, dealID, dir, size, item.Position.Level)

	guard := money.CloseGuard{Close: func() error {
		_, err := closeCanary(ctx, ex, ks, jw, epic, dealID, size, dir)
		return err
	}}
	defer func() {
		if err := guard.Ensure(); err != nil {
			logger.Error("CALIB CLOSE FAIL (%v) — HALT_NEW_ORDERS", err)
			_ = ks.HaltPersist()
			res = money.CalibrationResult{Failure: money.FailPositionReconcile}
			return
		}
		if !proveFlat(ctx, client, epic, dealID) {
			logger.Error("CALIB RECONCILE FAIL gold still open — HALT_NEW_ORDERS")
			_ = ks.HaltPersist()
			res = money.CalibrationResult{Failure: money.FailPositionReconcile}
		}
	}()

	accCur := spec.Currency
	if ar, aerr := client.GetAccounts(ctx); aerr == nil && session != nil {
		ident := execacct.Resolve(ar.Accounts, strings.TrimSpace(os.Getenv("AURUMFLOW_DEMO_ACCOUNT_ID")), "")
		if ident.Currency != "" {
			accCur = ident.Currency
		} else {
			for _, a := range ar.Accounts {
				if a.AccountID == session.CurrentAccountID && a.Currency != "" {
					accCur = a.Currency
					break
				}
			}
		}
	}
	maxD := time.Duration(calibSeconds) * time.Second
	if maxD < money.DefaultCalibMin {
		maxD = money.DefaultCalibMin
	}
	if maxD > 180*time.Second {
		maxD = 180 * time.Second
	}
	cfgEst := money.DefaultCalibrationConfig()
	cfgEst.TickSize = spec.TickSize
	cfgEst.Spread = details.Snapshot.Offer - details.Snapshot.Bid
	cfgEst.MetadataMPU = spec.MoneyPerPriceUnit

	_ = os.MkdirAll(filepath.Join("journals", "calibration"), 0o755)
	jf, _ := os.OpenFile(filepath.Join("journals", "calibration", strings.ToUpper(epic)+".jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if jf != nil {
		defer func() { _ = jf.Close() }()
	}

	var samples []money.CalibrationSample
	started := time.Now().UTC()
	lastLog := started
	ticker := time.NewTicker(money.DefaultSampleInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			res = money.EstimateTheilSen(samples, cfgEst)
			if res.Failure == "" && !res.OK {
				res.Failure = money.FailBrokerUncertain
			}
			return res
		case now := <-ticker.C:
			now = now.UTC()
			q, qerr := client.GetMarketDetails(ctx, epic)
			p, perr := client.GetPosition(ctx, dealID)
			if qerr != nil || perr != nil || q == nil || p == nil || p.Position.DealID != dealID {
				continue
			}
			if p.Position.Size != size {
				logger.Error("CALIB size changed dealId=%s", dealID)
				res = money.CalibrationResult{Failure: money.FailBrokerUncertain}
				return res
			}
			bid, ask := q.Snapshot.Bid, q.Snapshot.Offer
			s := money.CalibrationSample{
				Timestamp: now, DealID: dealID, Direction: dir, Size: size,
				Bid: bid, Ask: ask, CloseablePrice: money.CloseablePrice(dir, bid, ask),
				BrokerUPL: money.BrokerUPL(p.Position.Upnl, p.Position.ProfitLoss),
				PositionOpenLevel: p.Position.Level, QuoteAge: time.Since(now),
				PositionAge: now.Sub(started), AccountCurrency: accCur,
				EventTime: now, ReceiveTime: time.Now().UTC(),
			}
			samples = append(samples, s)
			if jf != nil {
				line, _ := json.Marshal(map[string]any{
					"instrument": epic, "dealId": dealID, "direction": dir, "size": size,
					"bid": bid, "ask": ask, "closeablePrice": s.CloseablePrice,
					"upl": s.BrokerUPL, "event_time": s.EventTime, "receive_time": s.ReceiveTime,
					"account_currency": accCur,
				})
				_, _ = jf.Write(append(line, '\n'))
			}
			if now.Sub(lastLog) >= 15*time.Second {
				est := money.EstimateTheilSen(samples, cfgEst)
				logger.Info("CALIB n=%d bid=%.2f ask=%.2f upl=%.4f pxRange=%.4f uplRange=%.4f elapsed=%s",
					len(samples), bid, ask, s.BrokerUPL, est.PriceRange, est.UPLRange, now.Sub(started).Round(time.Second))
				lastLog = now
			}
			if now.Sub(started) >= money.DefaultCalibMin {
				est := money.EstimateTheilSen(samples, cfgEst)
				if est.OK {
					res = est
					return res
				}
			}
			if now.Sub(started) >= maxD {
				res = money.EstimateTheilSen(samples, cfgEst)
				if res.Failure == "" && !res.OK {
					res.Failure = money.FailInsufficientSamples
				}
				return res
			}
		}
	}
}

func proveFlat(ctx context.Context, client *market.Client, epic, dealID string) bool {
	pr, err := client.GetPositions(ctx)
	if err != nil {
		return false
	}
	for _, p := range pr.Positions {
		if p.Position.DealID == dealID || strings.EqualFold(p.GetEpic(), epic) {
			return false
		}
	}
	return true
}

func demoWeekEligible(spec money.MonetaryInstrumentSpec, goldOpen int) bool {
	return spec.ValidationStatus == money.RuntimeValidated && spec.MoneyPerPriceUnit > 0 && goldOpen == 0
}
