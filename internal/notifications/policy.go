package notifications

import (
	"aurumflow/config"
	"time"
)

// Policy applies config-driven rules: category on/off, SWEEP filter, priority, throttle, dedup key.
type Policy struct {
	cfg *config.NotificationsConfig
}

// NewPolicy returns a policy for the given notifications config. Never returns nil; when cfg is nil, all checks return false.
func NewPolicy(cfg *config.NotificationsConfig) *Policy {
	return &Policy{cfg: cfg}
}

// CategoryEnabled returns true if the category is enabled in config.
func (p *Policy) CategoryEnabled(category string) bool {
	if p == nil || p.cfg == nil || p.cfg.Categories == nil {
		return false
	}
	switch category {
	case CategoryHealth:
		return p.cfg.Categories.Health
	case CategoryMarket:
		return p.cfg.Categories.Market
	case CategorySignals:
		return p.cfg.Categories.Signals
	case CategoryExecution:
		return p.cfg.Categories.Execution
	case CategoryRisk:
		return p.cfg.Categories.Risk
	case CategorySummary:
		return p.cfg.Categories.Summary
	default:
		return false
	}
}

// ShouldNotify returns true if the event's category is enabled. Caller ensures token/user present when client is active.
func (p *Policy) ShouldNotify(ev NotifEvent) bool {
	if p == nil {
		return false
	}
	return p.CategoryEnabled(ev.Category)
}

// SweepConfirmedShouldNotify returns true only if (a) strength >= SweepMinStrength, or (b) zoneBuy != 0 or zoneSell != 0, or (c) type changed BUY_SIDE <-> SELL_SIDE.
// Prevents "zoneBuy=0 zoneSell=0" spam.
func (p *Policy) SweepConfirmedShouldNotify(ev NotifEvent) bool {
	if !p.ShouldNotify(ev) {
		return false
	}
	minStrength := 0.0
	if p.cfg != nil && p.cfg.Thresholds != nil && p.cfg.Thresholds.SweepMinStrength > 0 {
		minStrength = p.cfg.Thresholds.SweepMinStrength
	}
	if strength, ok := ev.Payload["strength"].(float64); ok && strength >= minStrength {
		return true
	}
	if zoneBuy, ok := ev.Payload["zoneBuy"].(float64); ok && zoneBuy != 0 {
		return true
	}
	if zoneSell, ok := ev.Payload["zoneSell"].(float64); ok && zoneSell != 0 {
		return true
	}
	// type changed: store keeps last sweep type; caller can pass typeChanged in payload
	if typeChanged, ok := ev.Payload["typeChanged"].(bool); ok && typeChanged {
		return true
	}
	return false
}

// Priority returns Pushover priority: 0=INFO, 1=IMPORTANT/CRITICAL, 2=emergency only for POSITION_OPENED (live), DAILY_DD_UPDATE 80%/100%, DATA_STALE prolonged.
func (p *Policy) Priority(ev NotifEvent, apiMode string) int {
	switch ev.Severity {
	case SeverityInfo:
		return 0
	case SeverityImportant, SeverityCritical:
		// Emergency (2) only for specific cases
		if ev.Type == TypePositionOpened && apiMode == config.ModeLive {
			return 2
		}
		if ev.Type == TypeDailyDDUpdate {
			if ddPct, ok := ev.Payload["dd_pct"].(float64); ok {
				if limit, ok := ev.Payload["limit"].(float64); ok && limit > 0 {
					if ddPct >= limit*0.8 {
						return 2
					}
				}
			}
		}
		if ev.Type == TypeDataStale {
			if prolonged, ok := ev.Payload["prolonged"].(bool); ok && prolonged {
				return 2
			}
		}
		return 1
	default:
		return 0
	}
}

// MinInterval returns the minimum seconds between sends for this event type; 0 = no throttle.
func (p *Policy) MinInterval(evType string) time.Duration {
	if p == nil || p.cfg == nil || p.cfg.RateLimit == nil || p.cfg.RateLimit.MinIntervalSecondsByEvent == nil {
		return 0
	}
	sec, ok := p.cfg.RateLimit.MinIntervalSecondsByEvent[evType]
	if !ok || sec <= 0 {
		return 0
	}
	return time.Duration(sec) * time.Second
}

// DedupWindow returns the dedup window duration.
func (p *Policy) DedupWindow() time.Duration {
	if p == nil || p.cfg == nil || p.cfg.RateLimit == nil {
		return 900 * time.Second
	}
	if p.cfg.RateLimit.DedupWindowSeconds <= 0 {
		return 900 * time.Second
	}
	return time.Duration(p.cfg.RateLimit.DedupWindowSeconds) * time.Second
}

// AggregateWindow returns the aggregation window for SIGNAL_REJECTED etc.
func (p *Policy) AggregateWindow() time.Duration {
	if p == nil || p.cfg == nil || p.cfg.RateLimit == nil {
		return 1800 * time.Second
	}
	if p.cfg.RateLimit.AggregateWindowSeconds <= 0 {
		return 1800 * time.Second
	}
	return time.Duration(p.cfg.RateLimit.AggregateWindowSeconds) * time.Second
}

// DedupKey returns a key for deduplication: Type + Instrument + reason/session/timeframe.
func (p *Policy) DedupKey(ev NotifEvent) string {
	parts := []string{ev.Type, ev.Instrument, ev.Session}
	if ev.Type == TypeSignalRejected {
		if reason, ok := ev.Payload["reason_code"].(string); ok {
			parts = append(parts, reason)
		}
	}
	if ev.Type == TypeSweepConfirmed {
		if sweepType, ok := ev.Payload["sweep_type"].(string); ok {
			parts = append(parts, sweepType)
		}
	}
	return DedupKeyHash(parts...)
}
