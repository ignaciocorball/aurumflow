package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"aurumflow/config"
	"aurumflow/internal/capitalhist"
	"aurumflow/internal/demomut"
	"aurumflow/internal/execacct"
	"aurumflow/internal/logger"
	"aurumflow/internal/market"
	"aurumflow/internal/mktval"
	"aurumflow/internal/money"
	"aurumflow/internal/promote"
	"aurumflow/pkg/models"
)

var p89Spreads = map[string]float64{
	"GOLD": 0.50, "SILVER": 0.08, "OIL_CRUDE": 0.045, "US100": 1.80,
	"US500": 0.60, "US30": 2.00, "DE40": 4.00, "UK100": 3.00, "J225": 10.00, "CN50": 10.00,
}

func runP810(ctx context.Context, statusAddr string, shadowMin int, stay bool) {
	if statusAddr == "" || statusAddr == "127.0.0.1:8765" {
		statusAddr = "127.0.0.1:8767"
	}
	rows := loadP89Evidence()
	by := map[string]promote.Evidence{}
	for i := range rows {
		by[rows[i].Market] = rows[i]
	}
	us := by["US100"]
	gold := by["GOLD"]
	logger.Info("P8.10 audit GOLD n=%d exp=%.3f stress=%.3f pf=%.2f dd=%.2f subs=%d",
		gold.HoldoutN, gold.NormalExp, gold.StressExp, gold.ProfitFactor, gold.MaxDD, gold.PositiveThirds)
	logger.Info("P8.10 audit US100 n=%d exp=%.3f stress=%.3f pf=%.2f dd=%.2f subs=%d hash=%s",
		us.HoldoutN, us.NormalExp, us.StressExp, us.ProfitFactor, us.MaxDD, us.PositiveThirds, prefix(us.DatasetHash, 16))

	spec, prevStatus := completeUS100Monetary(ctx)
	us.Monetary = spec.ValidationStatus
	us.BrokerSpec = "SPEC_READY"
	if goldSpec, err := money.LoadSpec(money.CachePath("", "GOLD")); err == nil {
		gold.Monetary = goldSpec.ValidationStatus
		gold.BrokerSpec = "SPEC_READY"
		gold.DemoStatus = "GOLD_STRATEGY"
		gold.LiveShadow = "RUNNING_8765"
		by["GOLD"] = gold
	}

	var ident execacct.Identity
	tradeable := false
	brokerOK := spec.DealingComplete()
	if sess, err := bootstrapDemoSession(ctx, config.ExecutionDisabled); err == nil {
		ident = sess.Identity
		if md, merr := sess.client.GetMarketDetails(ctx, "US100"); merr == nil && md != nil {
			tradeable = md.Instrument.Epic != "" && spec.DealingComplete()
			live := money.FromMarketDetails(md)
			if spec.Epic == "" {
				spec.Epic = live.Epic
			}
			brokerOK = live.DealingComplete()
			us.BrokerSpec = "SPEC_READY"
			logger.Info("P8.10 US100 broker=%s status=%s min=%.4f", execacct.Mask(ident.AccountID), md.Snapshot.MarketStatus, spec.MinDealSize)
		}
		if pos, perr := sess.client.GetPositions(ctx); perr == nil {
			left := countEpic(pos, "US100")
			spec.BrokerPositionsAfter = left
			spec.PositionClosed = left == 0
			logger.Info("P8.10 US100 broker positions=%d", left)
		}
	} else {
		logger.Warn("P8.10 DEMO session unavailable for live gates: %v", err)
	}

	in := promote.GateInput{
		Evidence: us, Normal: statsFromEvidence(us, false), Stress: statsFromEvidence(us, true),
		Account: ident, WantAccountID: strings.TrimSpace(os.Getenv("AURUMFLOW_DEMO_ACCOUNT_ID")),
		Env: "DEMO", MonetarySpec: spec, BrokerIdentity: "capital.com",
		BrokerUsable: brokerOK, TradeableCFD: tradeable || spec.DealingComplete(),
		DataQuality: us.DataQuality,
	}
	dec := promote.Evaluate(in)
	if dec.Pass {
		us.ResearchStatus = mktval.DemoEligible
		us.DemoStatus = "DEMO_ELIGIBLE"
		us.LiveShadow = "ARMED"
		logger.Info("P8.10 US100 DEMO_ELIGIBLE · DEMO_MIRROR_US100=ON")
	} else {
		us.DemoStatus = "SHADOW_VALIDATED"
		us.LiveShadow = "SHADOW"
		logger.Warn("P8.10 US100 remains SHADOW_VALIDATED failures=%v", dec.Failures)
	}
	us.Monetary = spec.ValidationStatus
	by["US100"] = us
	for i := range rows {
		if ev, ok := by[rows[i].Market]; ok {
			rows[i] = ev
		} else {
			rows[i].LiveShadow = "SHADOW"
			rows[i].DemoStatus = "NO"
			if rows[i].Monetary == "" {
				if cached, err := money.LoadSpec(money.CachePath("", rows[i].Market)); err == nil {
					rows[i].Monetary = cached.ValidationStatus
				} else {
					rows[i].Monetary = "UNVERIFIED"
				}
			}
			if rows[i].BrokerSpec == "" {
				rows[i].BrokerSpec = "SPEC_READY"
			}
		}
	}
	writeP810Report(rows, us, gold, spec, prevStatus, dec, ident)
	if !dec.Pass {
		logger.Info("P8.10 not arming US100 DEMO_MIRROR")
		return
	}
	runUS100Mirror(ctx, statusAddr, shadowMin, stay, spec, ident)
}

func loadP89Evidence() []promote.Evidence {
	var out []promote.Evidence
	root := filepath.Join("data", "research", "p89", "capital")
	for _, id := range mktval.Universe {
		m15 := loadP89Candles(root, id, "MINUTE_15")
		h1 := loadP89Candles(root, id, "HOUR")
		h4 := loadP89Candles(root, id, "HOUR_4")
		ev := promote.Extract(id, m15, h1, h4, p89Spreads[id])
		ev.BrokerIdentity = "capital.com"
		if cached, err := money.LoadSpec(money.CachePath("", id)); err == nil {
			ev.Monetary = cached.ValidationStatus
		}
		ev.BrokerSpec = "SPEC_READY"
		out = append(out, ev)
	}
	return out
}

func loadP89Candles(root, id, res string) []models.Candle {
	raw, err := os.ReadFile(filepath.Join(root, id, res+".json"))
	if err != nil {
		return nil
	}
	var cs []models.Candle
	if json.Unmarshal(raw, &cs) != nil {
		return nil
	}
	return capitalhist.DedupSort(cs)
}

func completeUS100Monetary(ctx context.Context) (money.MonetaryInstrumentSpec, string) {
	path := money.CachePath("", "US100")
	spec, err := money.LoadSpec(path)
	if err != nil {
		spec = money.MonetaryInstrumentSpec{Epic: "US100", Provider: "capital.com"}
	}
	prev := spec.ValidationStatus
	if money.EvidenceComplete(spec) {
		logger.Info("P8.10 US100 monetary already complete MPU=%.6f samples=%d", spec.MoneyPerPriceUnit, spec.CalibrationSamples)
		return spec, prev
	}
	samples, jerr := money.LoadCalibrationJournal(filepath.Join("journals", "calibration", "US100.jsonl"))
	if jerr == nil && len(samples) >= money.MinCalibSamples {
		cfg := money.DefaultCalibrationConfig()
		if spec.MoneyPerPriceUnit > 0 {
			cfg.MetadataMPU = spec.MoneyPerPriceUnit
		} else if spec.LotSize > 0 {
			cfg.MetadataMPU = spec.LotSize
		}
		res := money.EstimateTheilSen(samples, cfg)
		left := 0
		if sess, err := bootstrapDemoSession(ctx, config.ExecutionDisabled); err == nil {
			if pos, perr := sess.client.GetPositions(ctx); perr == nil {
				left = countEpic(pos, "US100")
			}
		}
		if res.OK {
			spec = money.CompleteEvidence(spec, res, left)
		} else {
			logger.Warn("P8.10 journal re-estimate failed: %s — keeping prior status=%s", res.Failure, spec.ValidationStatus)
			spec.PositionClosed = left == 0
			spec.BrokerPositionsAfter = left
		}
		if err := money.SaveSpec(path, spec); err != nil {
			logger.Warn("P8.10 persist US100 spec: %v", err)
		}
		logger.Info("P8.10 US100 monetary completed from P8.9 canary journal status=%s mpu=%.6f samples=%d px=%.4f upl=%.4f est=%s meta=%s closed=%v",
			spec.ValidationStatus, spec.MoneyPerPriceUnit, spec.CalibrationSamples, spec.PriceRange, spec.UPLRange, spec.Estimator, spec.MetadataStatus, spec.PositionClosed)
		if money.EvidenceComplete(spec) {
			return spec, prev
		}
	}
	logger.Warn("P8.10 US100 monetary incomplete after journal — running serialized P8.3 UPL-slope calibration")
	coord := demomut.New()
	want := strings.TrimSpace(os.Getenv("AURUMFLOW_DEMO_ACCOUNT_ID"))
	sess, err := bootstrapDemoSession(ctx, config.ExecutionDemo)
	if err != nil {
		logger.Error("P8.10 calibration blocked: %v", err)
		return spec, prev
	}
	if err := coord.Reserve(demomut.Request{Origin: demomut.OriginCalibration, Market: "US100", Env: "DEMO", Account: sess.Identity, WantAccountID: want}); err != nil {
		logger.Error("P8.10 calibration reserve: %v", err)
		return spec, prev
	}
	defer coord.Release()
	spec, _ = runUPLSlopeCalibration(ctx, "US100", spec)
	if pos, perr := sess.client.GetPositions(ctx); perr == nil {
		left := countEpic(pos, "US100")
		spec.BrokerPositionsAfter = left
		spec.PositionClosed = left == 0
	}
	_ = money.SaveSpec(path, spec)
	return spec, prev
}

func writeP810Report(rows []promote.Evidence, us, gold promote.Evidence, spec money.MonetaryInstrumentSpec, prev string, dec promote.Decision, ident execacct.Identity) {
	body := promote.GateTable(rows)
	body += "\n## GOLD exact evidence\n\n"
	body += evidenceBlock(gold)
	body += "\n## US100 exact evidence\n\n"
	body += evidenceBlock(us)
	body += "\n## US100 monetary\n\n"
	body += us100MoneyBlock(spec, prev)
	body += "\n## DEMO account\n\n"
	body += "- Environment: DEMO\n"
	body += "- Account: " + verifiedLabel(ident) + "\n"
	body += "- Masked: `" + execacct.Mask(ident.AccountID) + "`\n"
	body += "- LIVE usable: NO\n"
	if !dec.Pass {
		body += "\n## US100 promotion decision\n\nSHADOW_VALIDATED — " + strings.Join(dec.Failures, "; ") + "\n"
	} else {
		body += "\n## US100 promotion decision\n\nDEMO_ELIGIBLE — DEMO_MIRROR_US100=ON\n"
	}
	_ = os.MkdirAll("research/reports", 0o755)
	if err := os.WriteFile("research/reports/PROMOTION_GATE_P8_10.md", []byte(body), 0o644); err != nil {
		logger.Error("P8.10 write report: %v", err)
	}
}

func evidenceBlock(ev promote.Evidence) string {
	return strings.Join([]string{
		"- Market: " + ev.Market,
		"- Dataset hash: `" + ev.DatasetHash + "`",
		"- Holdout dates: " + promote.HoldoutDates(ev),
		"- Holdout n: " + itoa(ev.HoldoutN),
		"- Net expectancy after NORMAL costs: " + f3(ev.NormalExp),
		"- Net expectancy under STRESS costs: " + f3(ev.StressExp),
		"- Hit rate: " + f3(ev.Hit),
		"- Profit factor: " + f2(ev.ProfitFactor),
		"- Max drawdown (R): " + f2(ev.MaxDD),
		"- MFE: " + f3(ev.MFE),
		"- MAE: " + f3(ev.MAE),
		"- Long expectancy: " + f3(ev.LongExp),
		"- Short expectancy: " + f3(ev.ShortExp),
		"- Positive chronological subperiods: " + itoa(ev.PositiveThirds) + " / 3",
		"- Largest-trade contribution: " + f3(ev.LargestShare),
		"- Data-quality status: " + ev.DataQuality,
		"- Broker identity: " + ev.BrokerIdentity,
		"- Broker spec status: " + ev.BrokerSpec,
		"- Monetary status: " + ev.Monetary,
		"",
	}, "\n")
}

func us100MoneyBlock(spec money.MonetaryInstrumentSpec, prev string) string {
	return strings.Join([]string{
		"- Previous: " + prev,
		"- Current: " + spec.ValidationStatus,
		"- MPU: " + f6(spec.MoneyPerPriceUnit),
		"- Currency: " + spec.MoneyPerPriceUnitCurrency,
		"- Samples: " + itoa(spec.CalibrationSamples),
		"- Price range: " + f4(spec.PriceRange),
		"- UPL range: " + f4(spec.UPLRange),
		"- Estimator: " + spec.Estimator,
		"- Dispersion (slope MAD): " + f6(spec.SlopeMAD),
		"- Metadata comparison: " + spec.MetadataStatus,
		"- Position closed: " + boolYes(spec.PositionClosed),
		"- Broker positions after: " + itoa(spec.BrokerPositionsAfter),
		"",
	}, "\n")
}

func statsFromEvidence(ev promote.Evidence, stress bool) mktval.Stats {
	st := mktval.Stats{
		N: ev.HoldoutN, Expectancy: ev.NormalExp, PF: ev.ProfitFactor, MaxDD: ev.MaxDD,
		Hit: ev.Hit, MFE: ev.MFE, MAE: ev.MAE, LongExp: ev.LongExp, ShortExp: ev.ShortExp,
		PositiveThirds: ev.PositiveThirds, LargestShare: ev.LargestShare, OutlierDom: ev.OutlierDominated,
		First: ev.HoldoutStart, Last: ev.HoldoutEnd,
	}
	if stress {
		st.Expectancy = ev.StressExp
	}
	return st
}

func countEpic(pos *market.PositionsResponse, epic string) int {
	if pos == nil {
		return 0
	}
	n := 0
	for _, p := range pos.Positions {
		if strings.EqualFold(p.GetEpic(), epic) {
			n++
		}
	}
	return n
}

func prefix(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func verifiedLabel(id execacct.Identity) string {
	if id.Verified && id.MayTrade {
		return "VERIFIED"
	}
	return "NOT_VERIFIED"
}

func boolYes(v bool) string {
	if v {
		return "YES"
	}
	return "NO"
}

func f2(v float64) string { return fmt.Sprintf("%.2f", v) }
func f3(v float64) string { return fmt.Sprintf("%.3f", v) }
func f4(v float64) string { return fmt.Sprintf("%.4f", v) }
func f6(v float64) string { return fmt.Sprintf("%.6f", v) }
func itoa(n int) string    { return fmt.Sprintf("%d", n) }
