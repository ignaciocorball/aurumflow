package terminal

import (
	"fmt"
	"sort"
	"strings"
)

func BuildNarratives(s Snapshot) Narratives {
	return Narratives{
		World:     worldNarrative(s),
		Matters:   mattersNow(s),
		Execution: executionNarrative(s),
		Health:    healthNarrative(s),
	}
}

func worldNarrative(s Snapshot) string {
	var parts []string
	dis := strings.ToUpper(s.World.Dislocation)
	switch dis {
	case "EXTREME":
		parts = append(parts, "A severe cross-market dislocation is currently detected.")
	case "SEVERE":
		parts = append(parts, "A severe cross-market dislocation is currently detected.")
	case "ELEVATED":
		parts = append(parts, "Markets are showing elevated dislocation.")
	default:
		parts = append(parts, "Global markets are broadly stable.")
	}
	ranked := rankedBySalience(s.Markets)
	if len(ranked) > 0 && ranked[0].Salience >= 55 {
		parts = append(parts, fmt.Sprintf("%s shows the strongest current %s salience.", ranked[0].Label, trendSide(ranked[0].Trend)))
	}
	if len(ranked) >= 3 && ranked[1].Salience >= 50 && ranked[2].Salience >= 50 {
		parts = append(parts, fmt.Sprintf("%s and %s are also unusually active.", ranked[1].Label, ranked[2].Label))
	} else if len(ranked) >= 2 && ranked[1].Salience >= 50 {
		parts = append(parts, fmt.Sprintf("%s is the next most active market.", ranked[1].Label))
	}
	if dis == "" || dis == "NORMAL" {
		parts = append(parts, "No severe cross-market dislocation is detected.")
	}
	return strings.Join(parts, " ")
}

func mattersNow(s Snapshot) []string {
	var out []string
	ranked := rankedBySalience(s.Markets)
	if len(ranked) > 0 && ranked[0].Salience >= 55 {
		out = append(out, fmt.Sprintf("%s equities are showing the strongest %s salience.", regionPhrase(ranked[0].Region), trendSide(ranked[0].Trend)))
	}
	active := []string{}
	for _, m := range ranked {
		if m.Salience >= 50 && (len(active) < 2) && (len(ranked) == 0 || m.ID != ranked[0].ID) {
			active = append(active, m.Label)
		}
	}
	if len(active) == 2 {
		out = append(out, fmt.Sprintf("%s and %s remain unusually active.", active[0], active[1]))
	}
	for _, m := range s.Markets {
		if m.ID == "GOLD" && m.Attention >= 50 && m.Salience < 45 && (m.Setup == "NONE" || m.Setup == "NO_SETUP" || m.Setup == "") {
			out = append(out, "GOLD has strong structural evidence but no valid Legacy setup.")
		}
		if m.ID == "US100" && (strings.Contains(strings.ToUpper(m.DemoStatus), "ELIGIBLE") || m.Pilot) {
			out = append(out, "US100 is armed for the next NY-eligible DEMO signal.")
		}
	}
	open := 0
	for _, p := range s.Execution.Positions {
		if p.Open {
			open++
		}
	}
	if open == 0 && s.Portfolio.OpenRisk == 0 {
		out = append(out, "No broker-backed positions are open.")
	}
	if len(out) == 0 {
		out = append(out, "No material anomalies require attention.")
	}
	if len(out) > 5 {
		out = out[:5]
	}
	return out
}

func executionNarrative(s Snapshot) string {
	var parts []string
	open := 0
	for _, p := range s.Execution.Positions {
		if p.Open {
			open++
			parts = append(parts, fmt.Sprintf("%s %s is open (%s).", p.Market, strings.ToLower(p.Direction), Display(p.Kind)))
		}
	}
	if open == 0 {
		parts = append(parts, "No positions are open.")
	}
	for _, p := range s.Execution.Positions {
		if p.Market == "GOLD" && !p.Open {
			wait := strings.TrimSpace(p.Waiting)
			if wait == "" || strings.Contains(strings.ToUpper(wait), "WAIT") {
				parts = append(parts, "GOLD is waiting for a Legacy setup.")
			}
		}
		if p.Market == "US100" && !p.Open && (strings.EqualFold(p.Armed, "ON") || strings.Contains(strings.ToUpper(p.Eligible), "ELIGIBLE")) {
			parts = append(parts, "US100 is armed for the next NY-eligible DEMO signal.")
		}
	}
	if s.Portfolio.OpenRisk == 0 {
		parts = append(parts, "Current planned broker risk is zero.")
	} else {
		parts = append(parts, fmt.Sprintf("Current planned broker risk is $%.2f.", s.Portfolio.OpenRisk))
	}
	return strings.Join(parts, " ")
}

func healthNarrative(s Snapshot) string {
	var degraded []string
	offline := []string{}
	for _, h := range s.Health {
		switch strings.ToUpper(h.Status) {
		case "DEGRADED", "STALE":
			degraded = append(degraded, h.Label)
		case "OFFLINE":
			offline = append(offline, h.Label)
		}
	}
	if len(offline) == 0 && len(degraded) == 0 {
		return "Data sources are healthy."
	}
	var parts []string
	if len(offline) > 0 {
		parts = append(parts, "Offline: "+strings.Join(offline, ", ")+".")
	}
	if len(degraded) > 0 {
		parts = append(parts, "Degraded: "+strings.Join(degraded, ", ")+".")
	}
	parts = append(parts, "Stale panels remain visible and are labelled.")
	return strings.Join(parts, " ")
}

func rankedBySalience(ms []Market) []Market {
	out := append([]Market(nil), ms...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Salience == out[j].Salience {
			return out[i].Attention > out[j].Attention
		}
		return out[i].Salience > out[j].Salience
	})
	return out
}

func trendSide(trend string) string {
	u := strings.ToUpper(trend)
	if strings.Contains(u, "DOWN") {
		return "downside"
	}
	if strings.Contains(u, "UP") {
		return "upside"
	}
	return "price"
}

func regionPhrase(region string) string {
	switch strings.ToUpper(region) {
	case "JAPAN":
		return "Japanese"
	case "UNITED_STATES", "US":
		return "US"
	case "EUROPE":
		return "European"
	case "CHINA_HONG_KONG", "CHINA_HK":
		return "China / Hong Kong"
	default:
		return Display(region)
	}
}
