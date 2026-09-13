package ops

import (
	"fmt"
	"math"
	"strings"
)

func FormatPrice(ok bool, v float64, prec int) string {
	if !ok || v == 0 || math.IsNaN(v) || math.IsInf(v, 0) {
		return Missing
	}
	if prec < 0 {
		prec = 2
	}
	return groupInt(fmt.Sprintf("%.*f", prec, v))
}

func FormatPct(ok bool, v float64) string {
	if !ok || math.IsNaN(v) || math.IsInf(v, 0) {
		return Missing
	}
	return fmt.Sprintf("%+.2f%%", v*100)
}

func FormatMs(ok bool, v float64) string {
	if !ok || v <= 0 || math.IsNaN(v) {
		return Missing
	}
	return fmt.Sprintf("%.0f ms", v)
}

func FormatMoney(ok bool, v float64) string {
	if !ok {
		return Missing
	}
	return "$" + groupInt(fmt.Sprintf("%.2f", v))
}

func EmptyState(kind string) string {
	switch strings.ToUpper(strings.TrimSpace(kind)) {
	case "CLOSED":
		return "MARKET CLOSED"
	case "HISTORY":
		return "WAITING FOR HISTORY"
	case "TRADEABLE":
		return "AWAITING TRADEABLE"
	case "SIGNAL":
		return "NO SIGNAL YET"
	default:
		return "WAITING"
	}
}

func HealthTone(label string) string {
	s := strings.ToUpper(strings.TrimSpace(label))
	switch {
	case s == "HEALTHY" || s == "SYNCED" || s == "OK" || s == "LIVE":
		return "HEALTHY"
	case s == "DEGRADED" || s == "PARTIAL" || s == "STALE" || s == "PREOPEN" || s == "WARMING":
		return "DEGRADED"
	case s == "FAILED" || s == "UNSYNCED" || s == "UNUSABLE" || s == "ERROR":
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}

func MilestoneCaption() string {
	return "NOT VALIDATION THRESHOLDS"
}

func GateMark(ok bool) string {
	if ok {
		return "✓"
	}
	return "○"
}

func WhyLines(text string, max int) []string {
	if max <= 0 {
		max = 5
	}
	var out []string
	for _, ln := range strings.Split(text, "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		out = append(out, ln)
		if len(out) >= max {
			break
		}
	}
	return out
}

func groupInt(s string) string {
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	parts := strings.SplitN(s, ".", 2)
	n := parts[0]
	var b strings.Builder
	for i, c := range n {
		if i > 0 && (len(n)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	out := b.String()
	if len(parts) == 2 {
		out += "." + parts[1]
	}
	if neg {
		return "-" + out
	}
	return out
}
