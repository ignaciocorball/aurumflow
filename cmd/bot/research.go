package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"runtime"
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

func runFlowExhaustionV1(ctx context.Context, fromS, toS, role string) {
	specPath := "research/specs/FLOW_EXHAUSTION_V1.json"
	hash, raw, err := research.LoadSpec(specPath)
	if err != nil {
		logger.Error("spec: %v", err)
		os.Exit(1)
	}
	_ = raw
	discFrom := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	discTo := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	from := time.Date(2026, 3, 17, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 14, 0, 0, 0, 0, time.UTC)
	studyRole := research.RoleHoldout
	if role == "discovery" {
		from = discFrom
		to = time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
		studyRole = research.RoleDiscovery
	}
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
	if studyRole == research.RoleHoldout {
		if !research.HoldoutValid(from, to.Add(24*time.Hour-time.Second), discFrom) {
			logger.Error("holdout range overlaps or does not end before discovery %s", discFrom)
			os.Exit(1)
		}
		if research.OverlapsDiscovery(from, to.Add(24*time.Hour-time.Second), discFrom, discTo) {
			logger.Error("holdout overlaps DISCOVERY_DATASET")
			os.Exit(1)
		}
	}
	arch := binancehist.New("data", "BTCUSDT")
	var paths []string
	var sums []string
	var integ research.HoldoutIntegrity
	integ.Start = from.Format("2006-01-02")
	integ.End = to.Format("2006-01-02")
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
		sum, ok, cErr := arch.VerifyChecksum(ctx, day, p)
		if cErr != nil {
			logger.Warn("checksum %s: %v", day, cErr)
		} else if !ok {
			logger.Warn("checksum mismatch %s", day)
		}
		q, _ := arch.ScanDay(p)
		integ.EventCount += q.Rows
		integ.Duplicates += q.Duplicates
		integ.OutOfOrder += q.OutOfOrder
		integ.Invalid += q.Invalid
		sums = append(sums, sum)
		paths = append(paths, p)
		logger.Info("holdout day=%s rows=%d checksum_ok=%v", day, q.Rows, cErr == nil && ok)
	}
	integ.Days = len(paths)
	integ.Files = len(paths)
	integ.Checksums = sums
	integ.DatasetHash = research.DatasetHash(sums)
	if len(paths) == 0 {
		logger.Error("no holdout archives")
		os.Exit(1)
	}
	logger.Info("flow-exhaustion-v1 role=%s spec=%s days=%d hash=%s", studyRole, hash[:12], len(paths), integ.DatasetHash[:12])
	st, err := research.RunExhaustionFromZips(ctx, paths, "BTCUSDT", studyRole, hash, integ)
	if err != nil {
		logger.Error("study: %v", err)
		os.Exit(1)
	}
	if err := research.WriteExhaustionReports("research/reports", st); err != nil {
		logger.Error("report: %v", err)
		os.Exit(1)
	}
	if studyRole == research.RoleHoldout {
		_ = os.WriteFile("research/reports/EXTERNAL_HOLDOUT_1_INTEGRITY.json", mustJSON(integ), 0o644)
	}
	logger.Info("flow-exhaustion done role=%s events=%d exh_n=%d decision=%s runtime=%s", st.Role, st.TradeEvents, st.Exhaustion.N, st.Decision, st.Runtime)
}

func runExhaustionMechanism(ctx context.Context) {
	reg := research.DefaultDatasetRegistry()
	_ = reg.Save("research/specs/DATASET_REGISTRY.json")
	type spec struct {
		role, from, to, label string
	}
	jobs := []spec{
		{research.RoleDiscovery, "2026-06-15", "2026-09-11", "DISCOVERY"},
		{research.RoleHoldout, "2026-03-17", "2026-06-14", "EXTERNAL_HOLDOUT_1"},
	}
	var blocks []research.MechBlock
	started := time.Now()
	events := 0
	for _, j := range jobs {
		from, _ := time.Parse("2006-01-02", j.from)
		to, _ := time.Parse("2006-01-02", j.to)
		if err := reg.ForbidUntouched(from, to); err != nil {
			logger.Error("%v", err)
			os.Exit(1)
		}
		arch := binancehist.New("data", "BTCUSDT")
		var paths []string
		for _, day := range binancehist.DaysInclusive(from, to) {
			p := arch.DayPath(day)
			if _, err := os.Stat(p); err != nil {
				logger.Warn("missing %s (not downloading untouched extra data)", day)
				continue
			}
			paths = append(paths, p)
		}
		if len(paths) == 0 {
			logger.Error("no archives for %s", j.label)
			os.Exit(1)
		}
		if j.role == research.RoleHoldout {
			var headers, invalid, rows int
			arch2 := binancehist.New("data", "BTCUSDT")
			for _, day := range binancehist.DaysInclusive(from, to) {
				p := arch2.DayPath(day)
				qday, err := arch2.ScanDay(p)
				if err != nil {
					continue
				}
				headers += qday.Headers
				invalid += qday.Invalid
				rows += qday.Rows
			}
			raw, _ := json.MarshalIndent(map[string]any{
				"previous_invalid_rows": 90, "cause": "binance CSV header agg_trade_id counted as invalid (underscore hid 'aggtrade' substring)",
				"fixed": true, "headers": headers, "invalid_rows": invalid, "data_rows": rows, "files": len(paths),
			}, "", "  ")
			_ = os.WriteFile("research/reports/EXTERNAL_HOLDOUT_1_INTEGRITY_P52.json", append(raw, '\n'), 0o644)
			logger.Info("holdout integrity rescan headers=%d invalid=%d rows=%d", headers, invalid, rows)
		}
		m1, snaps, q, ev, err := research.RunTape(ctx, paths, "BTCUSDT")
		if err != nil {
			logger.Error("%v", err)
			os.Exit(1)
		}
		events += ev
		fused, _, _ := research.BuildFused(m1, snaps)
		blk := research.AnalyzeMechanism(j.role, j.from+"→"+j.to, fused, m1, snaps)
		blocks = append(blocks, blk)
		logger.Info("mechanism %s events=%d legacy=%d exh=%d central=%s invalid_price=%d", j.label, ev, blk.LegacyN, blk.ExhaustionN, blk.Central, q.InvalidPrice)
	}
	rep := research.MechanismReport{Discovery: blocks[0], Holdout: blocks[1]}
	rep.Events = events
	rep.Runtime = time.Since(started).String()
	if time.Since(started) > 0 {
		rep.EventsPerSec = float64(events) / time.Since(started).Seconds()
	}
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	rep.MemMB = float64(ms.Alloc) / 1024 / 1024
	keys := []string{"flow_magnitude", "flow_efficiency", "impact_failure", "cvd", "trade_velocity"}
	rep.Consistency = map[string]string{}
	for _, k := range keys {
		de := rep.Discovery.Exhaustion[k].Mean - rep.Discovery.Neutral[k].Mean
		he := rep.Holdout.Exhaustion[k].Mean - rep.Holdout.Neutral[k].Mean
		if (de > 0) == (he > 0) {
			rep.Consistency[k] = "CONSISTENT"
		} else if de == 0 || he == 0 {
			rep.Consistency[k] = "MIXED"
		} else {
			rep.Consistency[k] = "INCONSISTENT"
		}
	}
	yes := 0
	for _, c := range []string{rep.Discovery.Central, rep.Holdout.Central} {
		if c == "YES" {
			yes++
		}
	}
	rep.Overall = "INCONCLUSIVE"
	if yes == 2 {
		rep.Overall = "YES"
	} else if rep.Discovery.Central == "NO" && rep.Holdout.Central == "NO" {
		rep.Overall = "NO"
	}
	if err := research.WriteMechanismReports("research/reports", rep); err != nil {
		logger.Error("%v", err)
		os.Exit(1)
	}
	logger.Info("mechanism report overall=%s runtime=%s", rep.Overall, rep.Runtime)
}

func runEventResolution(ctx context.Context) {
	reg := research.DefaultDatasetRegistry()
	jobs := []struct{ role, from, to, label string }{
		{research.RoleDiscovery, "2026-06-15", "2026-09-11", "DISCOVERY"},
		{research.RoleHoldout, "2026-03-17", "2026-06-14", "EXTERNAL_HOLDOUT_1"},
	}
	var blocks []research.EventResBlock
	started := time.Now()
	events := 0
	for _, j := range jobs {
		from, _ := time.Parse("2006-01-02", j.from)
		to, _ := time.Parse("2006-01-02", j.to)
		if err := reg.ForbidUntouched(from, to); err != nil {
			logger.Error("%v", err)
			os.Exit(1)
		}
		arch := binancehist.New("data", "BTCUSDT")
		var paths []string
		for _, day := range binancehist.DaysInclusive(from, to) {
			p := arch.DayPath(day)
			if _, err := os.Stat(p); err != nil {
				continue
			}
			paths = append(paths, p)
		}
		m1, snaps, _, _, err := research.RunTape(ctx, paths, "BTCUSDT")
		if err != nil {
			logger.Error("%v", err)
			os.Exit(1)
		}
		fused, _, _ := research.BuildFused(m1, snaps)
		blk, ev, err := research.AnalyzeEventResolution(j.role, j.from+"→"+j.to, fused, paths)
		if err != nil {
			logger.Error("%v", err)
			os.Exit(1)
		}
		events += ev
		blocks = append(blocks, blk)
		logger.Info("event-resolution %s exh_1s=%.3f cont_1s=%.3f neu_1s=%.3f central=%s", j.label, blk.Exhaustion["1s"].Mean, blk.Continuation["1s"].Mean, blk.Neutral["1s"].Mean, blk.Central)
	}
	rep := research.EventResReport{Discovery: blocks[0], Holdout: blocks[1], Events: events, Runtime: time.Since(started).String()}
	if time.Since(started) > 0 {
		rep.EventsPerS = float64(events) / time.Since(started).Seconds()
	}
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	rep.MemMB = float64(ms.Alloc) / 1024 / 1024
	if err := research.WriteEventResReports("research/reports", rep); err != nil {
		logger.Error("%v", err)
		os.Exit(1)
	}
}

func mustJSON(v any) []byte {
	b, _ := json.MarshalIndent(v, "", "  ")
	return append(b, '\n')
}

func gitHead() string {
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}
