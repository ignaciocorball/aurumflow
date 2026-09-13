package ops

import (
	"math"
	"strconv"
	"strings"
)

const Missing = "—"

func DisplayFloat(ok bool, v float64) string {
	if !ok || math.IsNaN(v) || math.IsInf(v, 0) {
		return Missing
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func DisplayInt(ok bool, v int) string {
	if !ok {
		return Missing
	}
	return strconv.Itoa(v)
}

func GoldMarketLabel(status string) string {
	s := strings.ToUpper(strings.TrimSpace(status))
	if s == "" {
		return "WAITING"
	}
	return s
}

func GoldQuotesOK(status string, bid, ask float64) bool {
	s := strings.ToUpper(strings.TrimSpace(status))
	return (s == "TRADEABLE" || s == "ON") && bid > 0 && ask > 0
}

func LegacyLabel(dir int, strategy string) string {
	if dir > 0 || strings.EqualFold(strategy, "LONG") {
		return "LONG"
	}
	if dir < 0 || strings.EqualFold(strategy, "SHORT") {
		return "SHORT"
	}
	return "NONE"
}

func V1Rail(class string) string {
	c := strings.ToUpper(class)
	switch {
	case strings.Contains(c, "EXHAUSTION"):
		return "EXHAUSTION"
	case strings.Contains(c, "CONTINUATION"):
		return "CONTINUATION"
	default:
		return "NEUTRAL"
	}
}

func DecisionWhy(st Status) string {
	var lines []string
	switch V1Rail(st.LastV1Class) {
	case "EXHAUSTION":
		if st.DirectionalP < 0 {
			lines = append(lines, "POTENTIAL SELL EXHAUSTION")
		} else if st.DirectionalP > 0 {
			lines = append(lines, "POTENTIAL LONG EXHAUSTION")
		} else {
			lines = append(lines, "POTENTIAL FLOW EXHAUSTION")
		}
	case "CONTINUATION":
		lines = append(lines, "FLOW CONTINUATION")
	default:
		if st.LastV1Class != "" {
			lines = append(lines, st.LastV1Class)
		}
	}
	if st.LastLegacyDir > 0 {
		lines = append(lines, "Legacy structure remains LONG.")
	} else if st.LastLegacyDir < 0 {
		lines = append(lines, "Legacy structure remains SHORT.")
	}
	if st.AggSell > st.AggBuy && st.AggSell > 0 {
		lines = append(lines, "Short-side aggression is elevated.")
	} else if st.AggBuy > st.AggSell && st.AggBuy > 0 {
		lines = append(lines, "Long-side aggression is elevated.")
	}
	if st.FlowEfficiency < 0 && st.ImpactFailure > 0 {
		lines = append(lines, "Price impact efficiency is deteriorating.")
	}
	if st.BidRepl > st.AskRepl && st.BidRepl > 0 {
		lines = append(lines, "Bid liquidity is replenishing.")
	} else if st.AskRepl > st.BidRepl && st.AskRepl > 0 {
		lines = append(lines, "Ask liquidity is replenishing.")
	}
	if st.L2QuotesOK && st.Microprice > 0 && st.L2Mid > 0 {
		if st.Microprice > st.L2Mid {
			lines = append(lines, "Microprice is resisting lower prices.")
		} else if st.Microprice < st.L2Mid {
			lines = append(lines, "Microprice is resisting higher prices.")
		}
	}
	return strings.Join(lines, "\n")
}
