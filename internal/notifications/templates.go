package notifications

import (
	"fmt"
	"strings"
	"time"

	"aurumflow/config"
)

// BuildTitle and BuildMessage return Pushover title and message for an event. Use api.mode in text.
func BuildTitle(ev NotifEvent, cfg *config.Config) string {
	mode := ""
	if cfg != nil {
		mode = cfg.API.Mode
	}
	instrument := ev.Instrument
	if instrument == "" {
		instrument = "—"
	}
	switch ev.Type {
	case TypeBotStart:
		return "AurumFlow • STARTED"
	case TypeBotStop:
		return "AurumFlow • STOPPED"
	case TypeHeartbeat:
		return fmt.Sprintf("AurumFlow • %s", instrument)
	case TypeDataStale:
		tf, _ := ev.Payload["timeframe"].(string)
		return fmt.Sprintf("AurumFlow • DATA FROZEN %s", tf)
	case TypeSignalGenerated:
		dir, _ := ev.Payload["direction"].(string)
		session, _ := ev.Payload["session"].(string)
		return fmt.Sprintf("AurumFlow • SIGNAL %s %s (%s)", dir, instrument, session)
	case TypeSignalRejected:
		return "AurumFlow • Signals Rejected"
	case TypeOrderSent:
		dir, _ := ev.Payload["direction"].(string)
		return fmt.Sprintf("AurumFlow • ORDER SENT %s %s", dir, instrument)
	case TypeOrderRejected:
		return fmt.Sprintf("AurumFlow • ORDER REJECTED %s", instrument)
	case TypePositionOpened:
		dir, _ := ev.Payload["direction"].(string)
		return fmt.Sprintf("AurumFlow • OPENED %s %s", dir, instrument)
	case TypePositionClosed:
		return fmt.Sprintf("AurumFlow • CLOSED %s", instrument)
	case TypeDailyDDUpdate:
		return "AurumFlow • Daily DD Alert"
	case TypeTradingHalted:
		return "AurumFlow • TRADING HALTED"
	case TypeDroppedSummary:
		return "AurumFlow • Notifs Dropped"
	default:
		return fmt.Sprintf("AurumFlow • %s %s (mode=%s)", ev.Type, instrument, mode)
	}
}

// BuildMessage returns the message body for the event (compact, readable).
func BuildMessage(ev NotifEvent, cfg *config.Config) string {
	mode := "paper"
	if cfg != nil {
		mode = cfg.API.Mode
	}
	switch ev.Type {
	case TypeBotStart:
		version, _ := ev.Payload["version"].(string)
		return fmt.Sprintf("mode=%s | %s", mode, version)
	case TypeBotStop:
		return fmt.Sprintf("mode=%s", mode)
	case TypeHeartbeat:
		state, _ := ev.Payload["state"].(string)
		balance, _ := ev.Payload["balance"].(float64)
		openPos, _ := ev.Payload["open_positions"].(int)
		maxTrades := 0
		if cfg != nil {
			maxTrades = cfg.Risk.MaxTrades
		}
		dd, _ := ev.Payload["daily_dd"].(float64)
		marketStatus, _ := ev.Payload["market_status"].(string)
		lastDecision, _ := ev.Payload["last_decision"].(string)
		return fmt.Sprintf("state=%s balance=%.2f pos=%d/%d DD=%.2f%%\nmarket=%s | %s", state, balance, openPos, maxTrades, dd, marketStatus, lastDecision)
	case TypeDataStale:
		tf, _ := ev.Payload["timeframe"].(string)
		ageMin, _ := ev.Payload["age_minutes"].(float64)
		maxMin, _ := ev.Payload["max_age_minutes"].(float64)
		return fmt.Sprintf("%s age %.0fm > %.0fm | block_on_data_frozen=true\nTrading halted until data recovers", tf, ageMin, maxMin)
	case TypeSignalGenerated:
		entry, _ := ev.Payload["entry"].(float64)
		sl, _ := ev.Payload["sl"].(float64)
		tp, _ := ev.Payload["tp"].(float64)
		score, _ := ev.Payload["score"].(int)
		th := 6
		if cfg != nil {
			th = cfg.Strategy.ScoreThreshold
		}
		conf, _ := ev.Payload["confidence"].(float64)
		trendM15, _ := ev.Payload["structure_m15"].(string)
		trendH1, _ := ev.Payload["trend_h1"].(string)
		rsi, _ := ev.Payload["rsi"].(float64)
		atr, _ := ev.Payload["atr"].(float64)
		return fmt.Sprintf("Entry %.2f | SL %.2f | TP %.2f\nScore %d (th=%d) | Conf %.2f\nM15 %s | H1 %s | RSI %.1f | ATR %.2f", entry, sl, tp, score, th, conf, trendM15, trendH1, rsi, atr)
	case TypeSignalRejected:
		// Aggregated template: "Signals generated: N | Rejected (reason): M | Last: dir score=X conf=Y"
		total, _ := ev.Payload["total"].(int)
		byReason, _ := ev.Payload["by_reason"].(map[string]int)
		lastDir, _ := ev.Payload["last_direction"].(string)
		lastScore, _ := ev.Payload["last_score"].(int)
		lastConf, _ := ev.Payload["last_conf"].(float64)
		var parts []string
		if byReason != nil {
			for reason, count := range byReason {
				parts = append(parts, fmt.Sprintf("%s: %d", reason, count))
			}
		}
		return fmt.Sprintf("Total %d | %s\nLast: %s score=%d conf=%.2f", total, strings.Join(parts, " | "), lastDir, lastScore, lastConf)
	case TypeOrderSent:
		dir, _ := ev.Payload["direction"].(string)
		size, _ := ev.Payload["size"].(float64)
		entry, _ := ev.Payload["entry"].(float64)
		sl, _ := ev.Payload["sl"].(float64)
		tp, _ := ev.Payload["tp"].(float64)
		return fmt.Sprintf("%s size=%.2f entry=%.2f SL=%.2f TP=%.2f", dir, size, entry, sl, tp)
	case TypeOrderRejected:
		errMsg, _ := ev.Payload["error"].(string)
		return fmt.Sprintf("error: %s", errMsg)
	case TypePositionOpened:
		dir, _ := ev.Payload["direction"].(string)
		avg, _ := ev.Payload["fill_price"].(float64)
		sl, _ := ev.Payload["sl"].(float64)
		tp, _ := ev.Payload["tp"].(float64)
		riskPct := 0.0
		if cfg != nil {
			riskPct = cfg.Risk.RiskPerTrade
		}
		openPos, _ := ev.Payload["open_positions"].(int)
		maxTrades := 0
		if cfg != nil {
			maxTrades = cfg.Risk.MaxTrades
		}
		return fmt.Sprintf("%s Avg %.2f | SL %.2f | TP %.2f | Risk %.2f%%\nMode: %s | Positions: %d/%d", dir, avg, sl, tp, riskPct, mode, openPos, maxTrades)
	case TypePositionClosed:
		reason, _ := ev.Payload["exit_reason"].(string)
		pnlPct, _ := ev.Payload["pnl_pct"].(float64)
		pnlR, _ := ev.Payload["pnl_r"].(float64)
		durMin, _ := ev.Payload["duration_minutes"].(int)
		return fmt.Sprintf("PnL %.2f%% | R %.2f | Reason %s\nDuration %d min", pnlPct, pnlR, reason, durMin)
	case TypeDailyDDUpdate:
		ddPct, _ := ev.Payload["dd_pct"].(float64)
		limit, _ := ev.Payload["limit"].(float64)
		action, _ := ev.Payload["action"].(string)
		return fmt.Sprintf("DD %.2f%% / limit %.2f%% | %s", ddPct, limit, action)
	case TypeTradingHalted:
		reason, _ := ev.Payload["reason"].(string)
		state, _ := ev.Payload["state"].(string)
		return fmt.Sprintf("reason: %s | state=%s", reason, state)
	case TypeDroppedSummary:
		dropped, _ := ev.Payload["dropped"].(int)
		return fmt.Sprintf("%d notifs dropped (queue full)", dropped)
	default:
		return fmt.Sprintf("%s at %s", ev.Type, ev.Timestamp.Format(time.RFC3339))
	}
}

// BuildTitleTelegram returns Telegram title with safe HTML; escape only variable values.
func BuildTitleTelegram(ev NotifEvent, cfg *config.Config) string {
	instrument := ev.Instrument
	if instrument == "" {
		instrument = "—"
	}
	instrument = EscapeHTMLValue(instrument)
	switch ev.Type {
	case TypeBotStart:
		return "<b>AurumFlow • STARTED</b>"
	case TypeBotStop:
		return "<b>AurumFlow • STOPPED</b>"
	case TypeHeartbeat:
		return "<b>AurumFlow • " + instrument + "</b>"
	case TypeDataStale:
		tf, _ := ev.Payload["timeframe"].(string)
		return "<b>AurumFlow • DATA FROZEN " + EscapeHTMLValue(tf) + "</b>"
	case TypeSignalGenerated:
		dir, _ := ev.Payload["direction"].(string)
		session, _ := ev.Payload["session"].(string)
		return "<b>AurumFlow • SIGNAL " + EscapeHTMLValue(dir) + " " + instrument + " (" + EscapeHTMLValue(session) + ")</b>"
	case TypeSignalRejected:
		return "<b>AurumFlow • Signals Rejected</b>"
	case TypeOrderSent:
		dir, _ := ev.Payload["direction"].(string)
		return "<b>AurumFlow • ORDER SENT " + EscapeHTMLValue(dir) + " " + instrument + "</b>"
	case TypeOrderRejected:
		return "<b>AurumFlow • ORDER REJECTED " + instrument + "</b>"
	case TypePositionOpened:
		dir, _ := ev.Payload["direction"].(string)
		return "<b>AurumFlow • OPENED " + EscapeHTMLValue(dir) + " " + instrument + "</b>"
	case TypePositionClosed:
		return "<b>AurumFlow • CLOSED " + instrument + "</b>"
	case TypeDailyDDUpdate:
		return "<b>AurumFlow • Daily DD Alert</b>"
	case TypeTradingHalted:
		return "<b>AurumFlow • TRADING HALTED</b>"
	case TypeDroppedSummary:
		return "<b>AurumFlow • Notifs Dropped</b>"
	default:
		return "<b>AurumFlow • " + EscapeHTMLValue(ev.Type) + " " + instrument + "</b>"
	}
}

// BuildMessageTelegram returns Telegram message body: HTML-safe, no balance/size/account_id (safe for public channel).
func BuildMessageTelegram(ev NotifEvent, cfg *config.Config) string {
	mode := "paper"
	if cfg != nil {
		mode = cfg.API.Mode
	}
	mode = EscapeHTMLValue(mode)
	switch ev.Type {
	case TypeBotStart:
		version, _ := ev.Payload["version"].(string)
		return "mode=" + mode + " | " + EscapeHTMLValue(version)
	case TypeBotStop:
		return "mode=" + mode
	case TypeHeartbeat:
		state, _ := ev.Payload["state"].(string)
		openPos, _ := ev.Payload["open_positions"].(int)
		maxTrades := 0
		if cfg != nil {
			maxTrades = cfg.Risk.MaxTrades
		}
		dd, _ := ev.Payload["daily_dd"].(float64)
		marketStatus, _ := ev.Payload["market_status"].(string)
		lastDecision, _ := ev.Payload["last_decision"].(string)
		return fmt.Sprintf("state=%s pos=%d/%d DD=%.2f%%\nmarket=%s | %s",
			EscapeHTMLValue(state), openPos, maxTrades, dd, EscapeHTMLValue(marketStatus), EscapeHTMLValue(lastDecision))
	case TypeDataStale:
		tf, _ := ev.Payload["timeframe"].(string)
		ageMin, _ := ev.Payload["age_minutes"].(float64)
		maxMin, _ := ev.Payload["max_age_minutes"].(float64)
		return fmt.Sprintf("%s age %.0fm &gt; %.0fm | block_on_data_frozen=true\nTrading halted until data recovers",
			EscapeHTMLValue(tf), ageMin, maxMin)
	case TypeSignalGenerated:
		entry, _ := ev.Payload["entry"].(float64)
		sl, _ := ev.Payload["sl"].(float64)
		tp, _ := ev.Payload["tp"].(float64)
		score, _ := ev.Payload["score"].(int)
		th := 6
		if cfg != nil {
			th = cfg.Strategy.ScoreThreshold
		}
		conf, _ := ev.Payload["confidence"].(float64)
		trendM15, _ := ev.Payload["structure_m15"].(string)
		trendH1, _ := ev.Payload["trend_h1"].(string)
		rsi, _ := ev.Payload["rsi"].(float64)
		atr, _ := ev.Payload["atr"].(float64)
		return fmt.Sprintf("Entry %.2f | SL %.2f | TP %.2f\nScore %d (th=%d) | Conf %.2f\nM15 %s | H1 %s | RSI %.1f | ATR %.2f",
			entry, sl, tp, score, th, conf, EscapeHTMLValue(trendM15), EscapeHTMLValue(trendH1), rsi, atr)
	case TypeSignalRejected:
		total, _ := ev.Payload["total"].(int)
		byReason, _ := ev.Payload["by_reason"].(map[string]int)
		lastDir, _ := ev.Payload["last_direction"].(string)
		lastScore, _ := ev.Payload["last_score"].(int)
		lastConf, _ := ev.Payload["last_conf"].(float64)
		var parts []string
		for reason, count := range byReason {
			parts = append(parts, EscapeHTMLValue(reason)+": "+fmt.Sprintf("%d", count))
		}
		return fmt.Sprintf("Total %d | %s\nLast: %s score=%d conf=%.2f",
			total, strings.Join(parts, " | "), EscapeHTMLValue(lastDir), lastScore, lastConf)
	case TypeOrderSent:
		dir, _ := ev.Payload["direction"].(string)
		entry, _ := ev.Payload["entry"].(float64)
		sl, _ := ev.Payload["sl"].(float64)
		tp, _ := ev.Payload["tp"].(float64)
		return fmt.Sprintf("%s entry=%.2f SL=%.2f TP=%.2f", EscapeHTMLValue(dir), entry, sl, tp)
	case TypeOrderRejected:
		errMsg, _ := ev.Payload["error"].(string)
		return "error: " + EscapeHTMLValue(errMsg)
	case TypePositionOpened:
		dir, _ := ev.Payload["direction"].(string)
		avg, _ := ev.Payload["fill_price"].(float64)
		sl, _ := ev.Payload["sl"].(float64)
		tp, _ := ev.Payload["tp"].(float64)
		openPos, _ := ev.Payload["open_positions"].(int)
		maxTrades := 0
		if cfg != nil {
			maxTrades = cfg.Risk.MaxTrades
		}
		return fmt.Sprintf("%s Avg %.2f | SL %.2f | TP %.2f\nPositions: %d/%d",
			EscapeHTMLValue(dir), avg, sl, tp, openPos, maxTrades)
	case TypePositionClosed:
		reason, _ := ev.Payload["exit_reason"].(string)
		pnlPct, _ := ev.Payload["pnl_pct"].(float64)
		pnlR, _ := ev.Payload["pnl_r"].(float64)
		durMin, _ := ev.Payload["duration_minutes"].(int)
		return fmt.Sprintf("PnL %.2f%% | R %.2f | Reason %s\nDuration %d min",
			pnlPct, pnlR, EscapeHTMLValue(reason), durMin)
	case TypeDailyDDUpdate:
		ddPct, _ := ev.Payload["dd_pct"].(float64)
		limit, _ := ev.Payload["limit"].(float64)
		action, _ := ev.Payload["action"].(string)
		return fmt.Sprintf("DD %.2f%% / limit %.2f%% | %s", ddPct, limit, EscapeHTMLValue(action))
	case TypeTradingHalted:
		reason, _ := ev.Payload["reason"].(string)
		state, _ := ev.Payload["state"].(string)
		return "reason: " + EscapeHTMLValue(reason) + " | state=" + EscapeHTMLValue(state)
	case TypeDroppedSummary:
		dropped, _ := ev.Payload["dropped"].(int)
		return fmt.Sprintf("%d notifs dropped (queue full)", dropped)
	default:
		return EscapeHTMLValue(ev.Type) + " at " + ev.Timestamp.Format(time.RFC3339)
	}
}
