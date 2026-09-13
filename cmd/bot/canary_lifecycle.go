package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aurumflow/config"
	"aurumflow/internal/execution"
	"aurumflow/internal/journal"
	"aurumflow/internal/killswitch"
	"aurumflow/internal/logger"
	"aurumflow/internal/market"
	"aurumflow/internal/risk"
)

const (
	canaryReason   = "CANARY_INFRASTRUCTURE_TEST"
	canaryDecision = "INFRASTRUCTURE_CANARY"
	canaryDir      = "BUY"
)

func runLifecycleCanary(ctx context.Context, epicHint string) {
	if !demoEnvConfigured() {
		logger.Error("DEMO_RUNTIME = BLOCKED_MISSING_DEMO_CREDENTIALS")
		os.Exit(2)
	}
	cfg, err := loadDemoCanaryConfig(config.ExecutionDemo)
	if err != nil {
		logger.Error("%v", err)
		os.Exit(1)
	}
	if cfg.API.Environment != string(config.APIEnvDemo) || cfg.ExecMode() != config.ExecutionDemo {
		logger.Error("FATAL: canary-lifecycle requires APIEnvironment=DEMO and ExecutionMode=DEMO")
		os.Exit(1)
	}

	ks := killswitch.New(false, killswitch.DefaultFile)
	if ks.HaltNewOrders() {
		logger.Error("PRECHECK FAIL: kill switch ON")
		os.Exit(1)
	}

	client := market.NewDemoClient()
	client.APIKey = cfg.API.APIKey
	if config.IsLiveCapitalHost(client.BaseURL) || !config.IsDemoCapitalHost(client.BaseURL) || cfg.API.BaseURL != config.DemoAPIHost {
		logger.Error("FATAL: LIVE broker access is disabled in P1.")
		os.Exit(1)
	}
	logger.Info("PRECHECK environment=%s host=%s execution_mode=%s kill_switch=OFF", cfg.API.Environment, client.BaseURL, cfg.ExecMode())

	session, err := client.CreateSession(ctx, cfg.API.Identifier, cfg.API.Password)
	if err != nil {
		logger.Error("AUTH: FAIL (%v)", err)
		os.Exit(1)
	}
	logger.Info("AUTH: PASS")

	ar, err := client.GetAccounts(ctx)
	if err != nil {
		logger.Error("ACCOUNT: FAIL (%v)", err)
		os.Exit(1)
	}
	acc, err := risk.SelectDemoAccount(ar.Accounts, session.CurrentAccountID)
	if err != nil {
		logger.Error("ACCOUNT: FAIL (%v)", err)
		os.Exit(1)
	}
	_ = os.Setenv("AURUMFLOW_DEMO_ACCOUNT_ID", acc.AccountID)
	if acc.AccountID != session.CurrentAccountID {
		if err := client.SwitchAccount(ctx, acc.AccountID); err != nil {
			logger.Error("ACCOUNT: FAIL switch (%v)", err)
			os.Exit(1)
		}
	}
	logger.Info("ACCOUNT: PASS id=%s type=%s currency=%s balance=%.2f available=%.2f",
		config.RedactAccountID(acc.AccountID), emptyDash(acc.AccountType), emptyDash(acc.Currency),
		acc.Balance.Balance, acc.Balance.Available)

	if strings.TrimSpace(epicHint) == "" {
		epicHint = os.Getenv("AURUMFLOW_EPIC")
	}
	epic, how, err := discoverTradeableEpic(ctx, client, epicHint)
	if err != nil {
		logger.Error("DISCOVERY: FAIL (%v)", err)
		os.Exit(1)
	}
	logger.Info("DISCOVERY: epic=%s via=%s", epic, how)

	details, err := client.GetMarketDetails(ctx, epic)
	if err != nil {
		logger.Error("MARKET DETAILS: FAIL (%v)", err)
		os.Exit(1)
	}
	status := details.Snapshot.MarketStatus
	if !marketAllowsExecution(status) {
		logger.Error("DEMO LIFECYCLE: BLOCKED_MARKET_CLOSED status=%s", status)
		os.Exit(3)
	}

	size, err := canaryDealSize(details.DealingRules.MinDealSize.Value, details.DealingRules.MinSizeIncrement.Value, details.DealingRules.MaxDealSize.Value)
	if err != nil {
		logger.Error("SIZE: FAIL (%v)", err)
		os.Exit(1)
	}
	logger.Info("CANARY epic=%s name=%s type=%s currency=%s bid=%.6f ask=%.6f spread=%.6f min=%.6f increment=%.6f size=%.6f status=%s lot=%.6f contract=%.6f scale=%.6f",
		details.Instrument.Epic, emptyDash(details.Instrument.Name), details.Instrument.Type, emptyDash(details.Instrument.Currency),
		details.Snapshot.Bid, details.Snapshot.Offer, details.Snapshot.Offer-details.Snapshot.Bid,
		details.DealingRules.MinDealSize.Value, details.DealingRules.MinSizeIncrement.Value, size, status,
		details.Instrument.LotSize, details.Instrument.ContractSize, details.Instrument.ScalingFactor)

	pos, err := client.GetPositions(ctx)
	if err != nil {
		logger.Error("POSITIONS: FAIL (%v)", err)
		os.Exit(1)
	}
	epicOpen := 0
	for _, p := range pos.Positions {
		if strings.EqualFold(p.GetEpic(), epic) {
			epicOpen++
		}
	}
	if epicOpen != 0 {
		logger.Error("PRECHECK FAIL: existing %s positions=%d (resolve before opening)", epic, epicOpen)
		os.Exit(1)
	}

	jw, err := journal.NewFileWriter(filepath.Join("journals", "canary-lifecycle.jsonl"))
	if err != nil {
		logger.Error("journal: %v", err)
		os.Exit(1)
	}
	defer func() { _ = jw.Close() }()
	life := func(e journal.Lifecycle) {
		e.Epic = epic
		e.Environment = "DEMO"
		e.ExecutionMode = string(config.ExecutionDemo)
		if e.Reason == "" {
			e.Reason = canaryReason
		}
		_ = jw.WriteLifecycle(e)
	}

	spread := details.Snapshot.Offer - details.Snapshot.Bid
	logger.Info("PRE-OPEN ts=%s status=%s bid=%.6f ask=%.6f spread=%.6f balance=%.2f available=%.2f open=%d mode=DEMO kill=OFF reason=%s decision=%s",
		time.Now().UTC().Format(time.RFC3339), status, details.Snapshot.Bid, details.Snapshot.Offer, spread,
		acc.Balance.Balance, acc.Balance.Available, epicOpen, canaryReason, canaryDecision)
	life(journal.Lifecycle{Event: journal.EventOrderIntent, Direction: canaryDir, Size: size, Entry: details.Snapshot.Offer, Reason: canaryReason})

	if fresh, ferr := client.GetMarketDetails(ctx, epic); ferr == nil && fresh != nil {
		details = fresh
		if !marketAllowsExecution(details.Snapshot.MarketStatus) {
			logger.Error("DEMO LIFECYCLE: BLOCKED_MARKET_CLOSED status=%s (recheck)", details.Snapshot.MarketStatus)
			os.Exit(3)
		}
		if sz, serr := canaryDealSize(details.DealingRules.MinDealSize.Value, details.DealingRules.MinSizeIncrement.Value, details.DealingRules.MaxDealSize.Value); serr == nil {
			size = sz
		}
	}
	ex := execution.NewExecutor(client, epic, details.DealingRules.MinDealSize.Value, details.DealingRules.MinSizeIncrement.Value)
	openAt := time.Now().UTC()
	openRes, err := ex.OpenInfrastructure(ctx, canaryDir, size)
	if err != nil {
		life(journal.Lifecycle{Event: journal.EventOrderRejected, Direction: canaryDir, Size: size, Error: err.Error(), Reason: canaryReason})
		logger.Error("OPEN: FAIL (%v) — not retrying POST", err)
		os.Exit(1)
	}
	life(journal.Lifecycle{Event: journal.EventOrderSubmitted, DealRef: openRes.DealReference, Direction: canaryDir, Size: size, Reason: canaryReason, Status: "SUBMITTED"})
	logger.Info("OPEN: submitted dealRef=%s", openRes.DealReference)

	conf, err := pollConfirm(ctx, ex, openRes.DealReference)
	if err != nil {
		life(journal.Lifecycle{Event: journal.EventOrderRejected, DealRef: openRes.DealReference, Error: err.Error(), Reason: canaryReason})
		logger.Error("CONFIRM: FAIL (%v) — resolving positions, not opening another", err)
		reportGoldPositions(ctx, client, epic)
		os.Exit(1)
	}
	if rejectedConfirm(conf) {
		life(journal.Lifecycle{Event: journal.EventOrderRejected, DealRef: conf.DealReference, DealID: conf.DealID, Status: conf.DealStatus, Error: conf.Reason, Reason: canaryReason})
		logger.Error("CONFIRM: REJECTED status=%s dealStatus=%s reason=%s", conf.Status, conf.DealStatus, conf.Reason)
		os.Exit(1)
	}
	openLatency := time.Since(openAt)
	life(journal.Lifecycle{Event: journal.EventOrderConfirmed, DealRef: conf.DealReference, DealID: conf.DealID, Direction: conf.Direction, Size: size, ConfirmedSize: conf.Size, FillPrice: conf.Level, Status: conf.Status, Reason: canaryReason})
	life(journal.Lifecycle{Event: journal.EventPositionOpen, DealRef: conf.DealReference, DealID: conf.DealID, Direction: conf.Direction, Size: conf.Size, FillPrice: conf.Level, Status: conf.Status, Reason: canaryReason})
	logger.Info("CONFIRM: PASS dealId=%s status=%s dealStatus=%s size=%.4f level=%.2f latency=%s",
		conf.DealID, conf.Status, conf.DealStatus, conf.Size, conf.Level, openLatency)

	item, err := findPosition(ctx, client, conf.DealID, epic)
	if err != nil {
		logger.Error("READ: FAIL (%v) — not opening another; attempting close of known dealId", err)
		if conf.DealID != "" {
			if _, cErr := closeCanary(ctx, ex, ks, jw, epic, conf.DealID, size, conf.Direction); cErr != nil {
				os.Exit(1)
			}
		}
		os.Exit(1)
	}
	logger.Info("POSITION READ: PASS dealId=%s epic=%s size=%.4f direction=%s level=%.2f",
		item.Position.DealID, item.GetEpic(), item.Position.Size, item.Position.Direction, item.Position.Level)

	closeRes, err := closeCanary(ctx, ex, ks, jw, epic, item.Position.DealID, item.Position.Size, item.Position.Direction)
	if err != nil {
		os.Exit(1)
	}

	final, err := client.GetPositions(ctx)
	if err != nil {
		logger.Error("FINAL POSITIONS: FAIL (%v)", err)
		os.Exit(1)
	}
	present := false
	goldLeft := 0
	for _, p := range final.Positions {
		if p.Position.DealID == item.Position.DealID {
			present = true
		}
		if strings.EqualFold(p.GetEpic(), epic) {
			goldLeft++
		}
	}
	if present || goldLeft != 0 {
		logger.Error("FINAL BROKER STATE FAIL canary_present=%v gold_open=%d total=%d", present, goldLeft, len(final.Positions))
		_ = ks.HaltPersist()
		os.Exit(1)
	}
	logger.Info("FINAL BROKER STATE: canary_present=false gold_open=0 total_open=%d close_level=%.2f pnl=%.4f latency=%s",
		len(final.Positions), closeRes.Level, closeRes.PnL, closeRes.Latency)
	logValuePerPointEvidence(details)
	logger.Info("DEMO lifecycle canary complete")
}

type closeOutcome struct {
	DealRef string
	Level   float64
	PnL     float64
	Latency time.Duration
}

func closeCanary(ctx context.Context, ex *execution.Executor, ks *killswitch.Switch, jw journal.JournalWriter, epic, dealID string, size float64, direction string) (closeOutcome, error) {
	life := func(e journal.Lifecycle) {
		e.Epic = epic
		e.Environment = "DEMO"
		e.ExecutionMode = string(config.ExecutionDemo)
		if e.Reason == "" {
			e.Reason = canaryReason
		}
		_ = jw.WriteLifecycle(e)
	}
	life(journal.Lifecycle{Event: journal.EventPositionCloseRequested, DealID: dealID, Direction: direction, Size: size, Reason: canaryReason})
	closeAt := time.Now().UTC()
	cl, err := ex.ClosePosition(ctx, dealID)
	if err != nil {
		life(journal.Lifecycle{Event: journal.EventPositionCloseRejected, DealID: dealID, Error: err.Error(), Reason: canaryReason})
		logger.Error("CLOSE: FAIL (%v) — activating HALT_NEW_ORDERS", err)
		if pErr := ks.HaltPersist(); pErr != nil {
			logger.Error("kill switch persist: %v", pErr)
		}
		logger.Error("recovery: DEMO flatten only — go run ./cmd/bot --flatten --flatten-confirm")
		return closeOutcome{}, err
	}
	latency := time.Since(closeAt)
	life(journal.Lifecycle{
		Event: journal.EventPositionClosed, DealRef: cl.DealReference, DealID: cl.DealID,
		Direction: direction, Size: size, ClosePrice: cl.Level, Status: cl.Status,
		CloseReason: canaryReason, Reason: canaryReason,
	})
	logger.Info("CLOSE: PASS dealRef=%s status=%s level=%.2f pnl=%.4f latency=%s", cl.DealReference, cl.Status, cl.Level, cl.PnL, latency)
	return closeOutcome{DealRef: cl.DealReference, Level: cl.Level, PnL: cl.PnL, Latency: latency}, nil
}

func pollConfirm(ctx context.Context, ex *execution.Executor, dealRef string) (*execution.ConfirmResult, error) {
	var last error
	for i := 0; i < 8; i++ {
		cr, err := ex.Confirm(ctx, dealRef)
		if err != nil {
			last = err
			time.Sleep(400 * time.Millisecond)
			continue
		}
		if cr != nil && (cr.DealID != "" || rejectedConfirm(cr) || strings.EqualFold(cr.DealStatus, "ACCEPTED") || strings.EqualFold(cr.Status, "OPEN")) {
			return cr, nil
		}
		last = err
		time.Sleep(400 * time.Millisecond)
	}
	if last != nil {
		return nil, last
	}
	return nil, context.DeadlineExceeded
}

func rejectedConfirm(cr *execution.ConfirmResult) bool {
	if cr == nil {
		return false
	}
	s := strings.ToUpper(cr.DealStatus + " " + cr.Status)
	return strings.Contains(s, "REJECT") || strings.Contains(s, "DELETED")
}

func findPosition(ctx context.Context, client *market.Client, dealID, epic string) (*market.PositionItem, error) {
	var last error
	for i := 0; i < 8; i++ {
		if dealID != "" {
			if one, err := client.GetPosition(ctx, dealID); err == nil && one != nil && one.Position.DealID != "" {
				return one, nil
			} else if err != nil {
				last = err
			}
		}
		pr, err := client.GetPositions(ctx)
		if err != nil {
			last = err
			time.Sleep(400 * time.Millisecond)
			continue
		}
		var matches []*market.PositionItem
		for i := range pr.Positions {
			p := &pr.Positions[i]
			if dealID != "" && p.Position.DealID == dealID {
				return p, nil
			}
			if strings.EqualFold(p.GetEpic(), epic) {
				matches = append(matches, p)
			}
		}
		if len(matches) == 1 {
			return matches[0], nil
		}
		if len(matches) > 1 {
			return nil, fmt.Errorf("ambiguous %s positions=%d after open; not guessing", epic, len(matches))
		}
		time.Sleep(400 * time.Millisecond)
	}
	if last != nil {
		return nil, last
	}
	return nil, errPositionNotFound(dealID, epic)
}

func errPositionNotFound(dealID, epic string) error {
	return fmt.Errorf("position not found dealId=%s epic=%s", dealID, epic)
}

func reportGoldPositions(ctx context.Context, client *market.Client, epic string) {
	pr, err := client.GetPositions(ctx)
	if err != nil {
		logger.Error("positions resolve failed: %v", err)
		return
	}
	n := 0
	for _, p := range pr.Positions {
		if strings.EqualFold(p.GetEpic(), epic) {
			n++
			logger.Info("open GOLD dealId=%s size=%.4f dir=%s", p.Position.DealID, p.Position.Size, p.Position.Direction)
		}
	}
	logger.Info("resolve: gold_open=%d total=%d", n, len(pr.Positions))
}

func logValuePerPointEvidence(details *market.MarketDetailsResponse) {
	if details == nil {
		logger.Info("VALUE_PER_POINT: UNKNOWN (no details)")
		return
	}
	logger.Info("VALUE_PER_POINT evidence lotSize=%.6f contractSize=%.6f scalingFactor=%.6f valueOfOnePip=%s marginFactor=%.6f currency=%s minStop=%.6f",
		details.Instrument.LotSize, details.Instrument.ContractSize, details.Instrument.ScalingFactor,
		strings.TrimSpace(string(details.Instrument.ValueOfOnePip)), details.Instrument.MarginFactor, details.Instrument.Currency,
		details.DealingRules.MinStopOrProfitDistance.Value)
	logger.Info("VALUE_PER_POINT: UNKNOWN (no canonical monetary field proven; not calibrated from a single trade)")
}
