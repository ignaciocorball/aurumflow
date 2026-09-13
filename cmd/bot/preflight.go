package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"aurumflow/config"
	"aurumflow/internal/binanceusdm"
	"aurumflow/internal/killswitch"
	"aurumflow/internal/logger"
	"aurumflow/internal/market"
	"aurumflow/internal/money"
	"aurumflow/internal/okxswap"
	"aurumflow/internal/ops"
	"aurumflow/internal/research"
	"aurumflow/internal/risk"
)

func runOpsPreflight(ctx context.Context) {
	in := ops.PreflightInput{
		Version: runtime.Version(),
		TestsOK: true,
		DemoHost: config.DemoAPIHost,
		V1Provider: "binance_usdm_public",
		JournalDir: "journals",
		JournalWritable: ops.Writable("journals"),
		ProspectiveWritable: ops.Writable(research.ProspectiveDir),
		DiskFreeMB: 1024,
		OpenPositions: 0,
	}
	ks := killswitch.New(false, killswitch.DefaultFile)
	in.KillSwitch = ks.HaltNewOrders()

	if demoEnvConfigured() {
		cfg, err := loadDemoCanaryConfig(config.ExecutionDisabled)
		if err != nil {
			in.LiveRequested = strings.Contains(strings.ToLower(err.Error()), "live")
		} else {
			client := market.NewDemoClient()
			client.APIKey = cfg.API.APIKey
			if config.IsLiveCapitalHost(client.BaseURL) {
				in.LiveRequested = true
			}
			in.DemoHost = client.BaseURL
			if session, err := client.CreateSession(ctx, cfg.API.Identifier, cfg.API.Password); err == nil {
				in.AccountOK = true
				if ar, err := client.GetAccounts(ctx); err == nil {
					if acc, err := risk.SelectDemoAccount(ar.Accounts, session.CurrentAccountID); err == nil {
						logger.Info("CAPITAL DEMO account=%s currency=%s", config.RedactAccountID(acc.AccountID), acc.Currency)
					}
				}
				if pos, err := client.GetPositions(ctx); err == nil && pos != nil {
					in.OpenPositions = len(pos.Positions)
				}
				if md, err := client.GetMarketDetails(ctx, "GOLD"); err == nil && md != nil {
					in.GoldStatus = md.Snapshot.MarketStatus
					logger.Info("GOLD market_status=%s bid=%.2f offer=%.2f", md.Snapshot.MarketStatus, md.Snapshot.Bid, md.Snapshot.Offer)
					obs := ops.GoldObservatory{
						MarketStatus:  md.Snapshot.MarketStatus,
						PositionsOK:   true,
						OpenPositions: in.OpenPositions,
					}
					obs.QuotesOK = ops.GoldQuotesOK(md.Snapshot.MarketStatus, md.Snapshot.Bid, md.Snapshot.Offer)
					if obs.QuotesOK {
						obs.Bid, obs.Ask = md.Snapshot.Bid, md.Snapshot.Offer
						obs.Spread = md.Snapshot.Offer - md.Snapshot.Bid
					}
					if spec, err := money.LoadSpec(money.CachePath("", "GOLD")); err == nil {
						obs.Validation = spec.ValidationStatus
					}
					_ = ops.WriteGoldObservatory("", obs)
				}
				if spec, err := money.LoadSpec(money.CachePath("", "GOLD")); err == nil {
					in.GoldMonetary = spec.ValidationStatus
				}
			}
		}
	} else {
		in.GoldStatus = ""
	}

	bn := binanceusdm.DiagnoseWSS(ctx)
	in.BTCHealth = bn.Status == "OPERATIONAL"
	ox := okxswap.Diagnose(ctx)
	if ox.Status == "OPERATIONAL" {
		in.L2Provider = okxswap.Name
	}
	pctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	pr := binanceusdm.ProbeIncrementalDepth(pctx, 12*time.Second)
	cancel()
	if pr.Deltas > 0 && pr.Synced {
		in.L2Provider = "binance_usdm_public"
		in.L2Synced = true
		in.L2Deltas = pr.Deltas
	} else if ox.Status == "OPERATIONAL" {
		in.L2Provider = okxswap.Name
		in.L2Deltas = 0
	}

	out := ops.EvaluatePreflight(in)
	raw, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(raw))
	logger.Info("ops-preflight %s", out.Result)
	if out.Result == ops.PreflightBlocked {
		os.Exit(2)
	}
}
