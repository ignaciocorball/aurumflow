package terminal

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ClosedTrade struct {
	At        time.Time
	Market    string
	Origin    string
	Scope     string
	Direction string
	PnL       float64
	R         float64
	Canary    bool
	Reason    string
	Win       bool
}

func LoadTrades(root string) []ClosedTrade {
	var out []ClosedTrade
	out = append(out, scanJournal(filepath.Join(root, "journals", "demo-week.jsonl"), "GOLD_STRATEGY", "BROKER_DEMO")...)
	out = append(out, scanJournal(filepath.Join(root, "journals", "canary-lifecycle.jsonl"), "CALIBRATION_CANARY", "BROKER_DEMO")...)
	out = append(out, scanJournal(filepath.Join(root, "journals", "us100-shadow.jsonl"), "DEMO_MIRROR", "SHADOW")...)
	out = append(out, scanTradeDir(filepath.Join(root, "research", "strategy-trades"))...)
	return out
}

func scanTradeDir(dir string) []ClosedTrade {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []ClosedTrade
	for _, e := range ents {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".jsonl") {
			continue
		}
		origin := "GOLD_STRATEGY"
		if strings.Contains(strings.ToLower(name), "us100") {
			origin = "DEMO_MIRROR"
		}
		out = append(out, scanJournal(filepath.Join(dir, name), origin, "BROKER_DEMO")...)
	}
	return out
}

func scanJournal(path, origin, scope string) []ClosedTrade {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []ClosedTrade
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var raw map[string]any
		if json.Unmarshal([]byte(line), &raw) != nil {
			continue
		}
		ev := strings.ToLower(asString(raw["event"]))
		if ev != "position_closed" && ev != "trade_closed" && ev != "closed" {
			continue
		}
		tr := ClosedTrade{
			At:        parseTime(firstString(raw, "ts", "t", "at", "time")),
			Market:    strings.ToUpper(firstString(raw, "epic", "market", "symbol")),
			Origin:    origin,
			Scope:     scope,
			Direction: strings.ToUpper(firstString(raw, "direction", "side")),
			PnL:       firstFloat(raw, "pnl", "realized_pnl", "profit", "pl"),
			R:         firstFloat(raw, "r", "r_multiple", "return_r"),
			Reason:    firstString(raw, "reason", "close_reason"),
		}
		if tr.Market == "" {
			tr.Market = "UNKNOWN"
		}
		if strings.Contains(strings.ToUpper(tr.Reason), "CANARY") || origin == "CALIBRATION_CANARY" {
			tr.Canary = true
			tr.Origin = "CALIBRATION_CANARY"
		}
		if strings.Contains(strings.ToUpper(firstString(raw, "origin")), "SHADOW") {
			tr.Scope = "SHADOW"
		}
		tr.Win = tr.PnL > 0 || tr.R > 0
		out = append(out, tr)
	}
	return out
}

func AggregatePnL(trades []ClosedTrade, scope, strategy string, now time.Time, equity float64) Portfolio {
	p := Portfolio{
		Equity: equity, Currency: "USD", Scope: scope,
		OpenRiskCap: 10, PilotCap: 5, AggregateCap: 10,
		PreciousCap: 5, USEquityCap: 5,
	}
	now = now.UTC()
	today := now.Format("2006-01-02")
	weekStart := now.AddDate(0, 0, -int(now.Weekday())).Format("2006-01-02")
	monthPrefix := now.Format("2006-01")
	days := map[string]*DayPnL{}
	var filtered []ClosedTrade
	for _, t := range trades {
		if !includeTrade(t, scope, strategy) {
			continue
		}
		filtered = append(filtered, t)
		d := t.At.UTC().Format("2006-01-02")
		row, ok := days[d]
		if !ok {
			row = &DayPnL{Date: d, Best: t.PnL, Worst: t.PnL}
			days[d] = row
		}
		row.Realized += t.PnL
		row.R += t.R
		row.Trades++
		if t.Win {
			row.Wins++
		} else {
			row.Losses++
		}
		if t.PnL > row.Best {
			row.Best = t.PnL
		}
		if t.PnL < row.Worst {
			row.Worst = t.PnL
		}
		p.AllTime += t.PnL
		p.AllTimeR += t.R
		p.TradeCount++
		if d == today {
			p.TodayPnL += t.PnL
			p.TodayR += t.R
		}
		if d >= weekStart {
			p.WeeklyTotal += t.PnL
		}
		if strings.HasPrefix(d, monthPrefix) {
			p.MonthlyTotal += t.PnL
		}
	}
	p.HasBrokerTrades = len(filtered) > 0 && scope == "BROKER_DEMO"
	p.Realized = p.AllTime
	p.Net = p.Realized + p.Unrealized
	if equity > 0 {
		start := equity - p.AllTime
		if start > 0 {
			p.ReturnPct = (p.AllTime / start) * 100
		}
	}
	for _, row := range days {
		p.Calendar = append(p.Calendar, *row)
	}
	sortDays(p.Calendar)
	curve := equity
	if p.AllTime != 0 {
		curve = equity
	}
	running := equity - p.AllTime
	peak := running
	maxDD := 0.0
	for _, row := range p.Calendar {
		running += row.Realized
		if running > peak {
			peak = running
		}
		if peak > 0 {
			dd := (peak - running) / peak * 100
			if dd > maxDD {
				maxDD = dd
			}
		}
		p.EquityCurve = append(p.EquityCurve, CurvePt{T: row.Date, V: running})
	}
	if len(p.EquityCurve) == 0 && equity > 0 {
		p.EquityCurve = []CurvePt{{T: today, V: equity}}
	}
	p.DrawdownPct = maxDD
	_ = curve
	return p
}

func includeTrade(t ClosedTrade, scope, strategy string) bool {
	if t.Canary && scope != "ALL" {
		return false
	}
	switch strings.ToUpper(scope) {
	case "BROKER_DEMO":
		if t.Scope != "BROKER_DEMO" || t.Canary {
			return false
		}
	case "SHADOW":
		if t.Scope != "SHADOW" {
			return false
		}
	case "RESEARCH":
		if t.Scope != "RESEARCH" {
			return false
		}
	}
	switch strings.ToUpper(strategy) {
	case "GOLD":
		return t.Market == "GOLD" || t.Origin == "GOLD_STRATEGY"
	case "US100":
		return t.Market == "US100" || t.Origin == "DEMO_MIRROR"
	}
	return true
}

func sortDays(days []DayPnL) {
	for i := 0; i < len(days); i++ {
		for j := i + 1; j < len(days); j++ {
			if days[j].Date < days[i].Date {
				days[i], days[j] = days[j], days[i]
			}
		}
	}
}
