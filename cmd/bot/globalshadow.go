package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aurumflow/config"
	"aurumflow/internal/demoexec"
	"aurumflow/internal/eligibility"
	"aurumflow/internal/instrument"
	"aurumflow/internal/logger"
	"aurumflow/internal/money"
	"aurumflow/internal/opportunity"
	"aurumflow/internal/orchestrator"
	"aurumflow/internal/portfoliorisk"
	"aurumflow/internal/researchopp"
	"aurumflow/internal/scanner"
	"aurumflow/internal/worlddomain"
	"aurumflow/internal/worldstate"
)

func runGlobalShadow(ctx context.Context) {
	now := time.Now().UTC()
	in, res := worldstate.LoadOfficialInput(ctx, now)
	ws := worldstate.At(now, in)
	ranks := opportunity.Rank(ws)
	ws = scanner.AttachOpportunity(ws, ranks)
	orch := orchestrator.New()
	risk := portfoliorisk.New()
	elig := eligibility.New()
	demo := demoexec.New()
	var maps []instrument.Mapping
	if demoEnvConfigured() {
		maps = discoverCapitalMarkets(ctx)
	} else {
		logger.Info("global-shadow: no DEMO credentials — analysis-only, no broker discovery")
	}
	for _, m := range maps {
		elig.Set(eligibility.Advance(eligibility.Entry{Canonical: m.Canonical, Epic: m.Epic, DemoHost: true}, money.MonetaryInstrumentSpec{Epic: m.Epic}, false, false))
	}
	var proposals []orchestrator.DecisionProposal
	watch, setup, cand, blocked := 0, 0, 0, 0
	for _, r := range ranks {
		if r.Tier != worlddomain.TierA && r.Tier != worlddomain.TierB && r.Tier != worlddomain.TierC && r.Tier != worlddomain.TierIgnore {
			continue
		}
		e := elig.Get(r.Market)
		pr := risk.Evaluate(nil, portfoliorisk.Position{Canonical: r.Market, Region: string(r.State.Region), Asset: string(r.State.AssetClass), RiskKnown: false})
		p := orch.ProposeFull(ws, r, "", "", e.Status, pr.Blocks)
		proposals = append(proposals, p)
		switch p.Decision {
		case "SETUP":
			setup++
		case "DEMO_CANDIDATE":
			cand++
		case "BLOCKED":
			blocked++
		default:
			watch++
		}
		raw, _ := json.Marshal(p)
		_ = researchopp.Record("journals", researchopp.Signal{
			ID: r.Market + "-" + now.Format("20060102150405"), T0: now,
			Proposal: raw, Ranking: jsonBytes(r), MarketState: jsonBytes(r.State),
		})
		acc := demo.Accept(p, false)
		if acc.Accepted {
			logger.Error("global-shadow must never accept execution")
		}
	}
	rep := map[string]any{
		"as_of": ws.AsOf, "valid": ws.Valid, "fetch": res,
		"markets_analyzed": len(proposals), "watch": watch, "setup": setup,
		"demo_candidate": cand, "blocked": blocked, "proposals": proposals,
		"broker_mutation": orch.CanMutateBroker(),
		"multi_market_demo": demo.Enabled,
	}
	_ = worldstate.WriteJSON(filepath.Join("research/reports", "MULTI_MARKET_SHADOW.json"), rep)
	_ = worldstate.WriteMD(filepath.Join("research/reports", "MULTI_MARKET_SHADOW.md"), shadowMD(rep, proposals))
	logger.Info("global-shadow valid=%s analyzed=%d watch=%d setup=%d candidate=%d blocked=%d mutate=%v multi_demo=%v",
		ws.Valid, len(proposals), watch, setup, cand, blocked, orch.CanMutateBroker(), demo.Enabled)
}

func discoverCapitalMarkets(ctx context.Context) []instrument.Mapping {
	sess, err := bootstrapDemoSession(ctx, config.ExecutionDisabled)
	if err != nil {
		logger.Warn("global-shadow discovery: %v", err)
		return nil
	}
	var hits []instrument.SearchHit
	for _, term := range instrument.SearchTerms() {
		ms, err := sess.client.SearchMarkets(ctx, term)
		if err != nil {
			continue
		}
		for _, m := range ms {
			hits = append(hits, instrument.SearchHit{Epic: m.Epic, Name: m.InstrumentName, Status: m.MarketStatus})
			logger.Info("DISCOVER %s epic=%s status=%s", term, m.Epic, m.MarketStatus)
		}
	}
	maps := instrument.MapHits(hits)
	_ = instrument.PersistMappings("data/world/normalized/capital/mappings.json", maps)
	return maps
}

func runPrepareMarket(ctx context.Context, canonical string, calibrate bool) {
	canonical = strings.ToUpper(strings.TrimSpace(canonical))
	if canonical == "" {
		logger.Error("prepare-market requires --prepare-market <canonical>")
		os.Exit(1)
	}
	epic := canonical
	if maps, err := instrument.LoadMappings("data/world/normalized/capital/mappings.json"); err == nil {
		for _, m := range maps {
			if strings.EqualFold(m.Canonical, canonical) && m.Epic != "" {
				epic = m.Epic
				break
			}
		}
	}
	if calibrate {
		logger.Info("CALIBRATE MARKET canonical=%s epic=%s DEMO-only explicit", canonical, epic)
	} else {
		logger.Info("PREPARE MARKET canonical=%s epic=%s READ-ONLY", canonical, epic)
	}
	runPrepareInstrument(ctx, epic, calibrate)
}

func jsonBytes(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func shadowMD(rep map[string]any, ps []orchestrator.DecisionProposal) string {
	var b strings.Builder
	b.WriteString("# MULTI-MARKET SHADOW\n\nNo orders. MULTI_MARKET_DEMO_EXECUTION=OFF.\n\n")
	for _, p := range ps {
		b.WriteString("- " + p.Market + " attention=" + fnum(p.AttentionScore) + " coverage=" + fnum(p.Coverage) + " → " + p.Decision + "\n")
	}
	_ = rep
	return b.String()
}

func fnum(v float64) string {
	return strings.TrimRight(strings.TrimRight(
		func() string {
			s := ""
			n := int(v*10 + 0.5)
			neg := n < 0
			if neg {
				n = -n
			}
			s = itoaBot(n/10) + "." + itoaBot(n%10)
			if neg {
				s = "-" + s
			}
			return s
		}(), "0"), ".")
}
