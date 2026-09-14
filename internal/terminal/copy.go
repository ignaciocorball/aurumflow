package terminal

import (
	"strings"
	"unicode"
)

var copyMap = map[string]string{
	"RUNTIME_VALIDATED":              "Runtime validated",
	"BROKER_METADATA_ONLY":           "Broker metadata only",
	"POLICY_ALL":                     "All sessions",
	"TRADEABLE":                      "Tradeable",
	"CLOSED":                         "Closed",
	"UNKNOWN":                        "Unknown",
	"HEALTHY":                        "Healthy",
	"DEGRADED":                       "Degraded",
	"OFFLINE":                        "Offline",
	"STALE":                          "Stale",
	"DEMO":                           "DEMO",
	"SHADOW":                         "SHADOW",
	"RESEARCH":                       "Research",
	"DEMO_MIRROR":                    "DEMO mirror",
	"GOLD_STRATEGY":                  "GOLD strategy",
	"CALIBRATION_CANARY":             "Calibration canary",
	"PARTIAL":                        "Partial",
	"ELIGIBLE":                       "Eligible",
	"TRADE_LIFECYCLE_PASS":           "Trade lifecycle pass",
	"DEMO_OPERATIONALLY_TRUSTED":     "DEMO operationally trusted",
	"DEMO_ELIGIBLE":                  "DEMO eligible",
	"DEMO_OPERATIONAL":               "DEMO operational",
	"DATA_READY":                     "Data ready",
	"RESEARCH_READY":                 "Research ready",
	"SHADOW_VALIDATED":               "Shadow validated",
	"RESEARCH_REJECTED":              "Research rejected",
	"BLOCKED":                        "Blocked",
	"ANALYSIS_ONLY":                  "Analysis only",
	"LIVE_PROHIBITED":                "Live prohibited",
	"NO_SETUP":                       "No setup",
	"NONE":                           "None",
	"INSUFFICIENT_DATA":              "Insufficient data",
	"CURRENT_WORLD_STATE_VALID":      "World current",
	"FIXTURE_CONTAMINATED":           "Fixture contaminated",
	"FIXTURE_OR_TEST":                "Fixture or test",
	"IMPOSSIBLE / FAIL-CLOSED":       "Live impossible / fail-closed",
	"SEEK_LIQUIDITY":                 "Seeking liquidity",
	"REJ_NO_SIGNAL":                  "No signal",
	"VALIDATED_EXTERNAL_HOLDOUT":     "Validated (external holdout)",
	"BOOK_CAPABILITY_LIMITED":        "Book capability limited",
	"CAPITAL_REST":                   "Capital REST",
	"OBSERVED":                       "Observed",
	"PROXY":                          "Proxy",
	"INFERRED":                       "Inferred",
	"DELAYED":                        "Delayed",
	"NORMAL":                         "Normal",
	"ELEVATED":                       "Elevated",
	"SEVERE":                         "Severe",
	"EXTREME":                        "Extreme",
	"RISK_ON":                        "Risk on",
	"RISK_OFF":                       "Risk off",
	"MIXED":                          "Mixed",
	"TRANSITION":                     "Transition",
	"STRONG_UP":                      "Strong up",
	"STRONG_DOWN":                    "Strong down",
	"UP":                             "Up",
	"DOWN":                           "Down",
	"FLAT":                           "Flat",
	"OIL_CRUDE":                      "OIL",
	"CHINA_HK":                       "China / Hong Kong",
}

func Display(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "—"
	}
	if v, ok := copyMap[s]; ok {
		return v
	}
	if strings.HasPrefix(s, "data_frozen:") {
		return strings.TrimPrefix(s, "data_frozen:") + " data temporarily stale"
	}
	if strings.HasPrefix(s, "rejected:") {
		return "Rejected: " + Display(strings.TrimPrefix(s, "rejected:"))
	}
	return humanizeToken(s)
}

func humanizeToken(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.Join(strings.Fields(s), " ")
	if s == "" {
		return "—"
	}
	runes := []rune(strings.ToLower(s))
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func MarketLabel(id string) string {
	switch strings.ToUpper(strings.TrimSpace(id)) {
	case "OIL_CRUDE", "OIL":
		return "OIL"
	case "CHINA_HK":
		return "China / HK"
	case "GOLD", "SILVER", "US100", "US500", "US30", "DE40", "UK100", "J225", "CN50", "BTC":
		return strings.ToUpper(id)
	default:
		return strings.ToUpper(strings.TrimSpace(id))
	}
}

func TrustHint(code string) string {
	switch strings.ToUpper(strings.TrimSpace(code)) {
	case "PARTIAL":
		return "Some operational gates are still incomplete. Observation is allowed; new risk is constrained."
	case "ELIGIBLE":
		return "Market and account gates allow DEMO consideration. Not a live-trading state."
	case "TRADE_LIFECYCLE_PASS":
		return "Open, protect, monitor, and close plumbing has passed DEMO checks."
	case "DEMO_OPERATIONALLY_TRUSTED":
		return "DEMO execution path is trusted for this pilot. Live trading remains impossible."
	default:
		return "Operational trust is a process-health label. It is not a performance score."
	}
}
