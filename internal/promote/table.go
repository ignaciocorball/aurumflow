package promote

import (
	"fmt"
	"strings"

)

func GateTable(rows []Evidence) string {
	var b strings.Builder
	b.WriteString("# PROMOTION GATE P8.10\n\n")
	b.WriteString("Frozen strategy: `LEGACY_NORMALIZED_V0` (`" + FrozenStrategy + "` / `" + "aurumflow/internal/legacynorm" + "`).\n\n")
	b.WriteString("Promotion uses holdout after NORMAL costs plus the STRESS catastrophic-reversal gate. Attention and Salience cannot promote or create orders.\n\n")
	b.WriteString("| Market | Research status | Holdout n | Net expectancy | PF | MaxDD | Cost stress | Subperiod robustness | Broker spec | Monetary | Live shadow | DEMO status |\n")
	b.WriteString("|---|---|---:|---:|---:|---:|---|---|---|---|---|---|\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "| %s | %s | %d | %.3f | %.2f | %.2f | %s | %s | %s | %s | %s | %s |\n",
			r.Market, r.ResearchStatus, r.HoldoutN, r.NormalExp, r.ProfitFactor, r.MaxDD,
			stressCell(r), subCell(r), empty(r.BrokerSpec), empty(r.Monetary), empty(r.LiveShadow), empty(r.DemoStatus))
	}
	b.WriteString("\n## Hard gates\n\n")
	b.WriteString("- holdout n ≥ 20\n")
	b.WriteString("- net holdout expectancy > 0 after NORMAL modeled costs\n")
	b.WriteString("- no catastrophic reversal under STRESS costs\n")
	b.WriteString("- profit factor > 1.1 (frozen research contract; also > 1)\n")
	b.WriteString("- max DD < 20R\n")
	b.WriteString("- ≥ 2 of 3 chronological holdout thirds non-negative\n")
	b.WriteString("- largest winner < 40% of total holdout R\n")
	b.WriteString("- data quality VALID\n")
	b.WriteString("- exact broker identity confirmed\n")
	b.WriteString("- broker spec usable\n")
	b.WriteString("- market tradeable as a CFD on the explicit DEMO account\n")
	b.WriteString("- monetary RUNTIME_VALIDATED with complete evidence\n")
	b.WriteString("- explicit DEMO account VERIFIED; LIVE rejected\n")
	return b.String()
}

func stressCell(r Evidence) string {
	if r.HoldoutN == 0 {
		return "n/a"
	}
	tag := "OK"
	if r.StressExp <= StressCatastrophic {
		tag = "FAIL"
	} else if r.NormalExp > 0 && r.StressExp < 0 {
		tag = "WEAK"
	}
	return fmt.Sprintf("%.3f (%s)", r.StressExp, tag)
}

func subCell(r Evidence) string {
	return fmt.Sprintf("%d/3", r.PositiveThirds)
}

func empty(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}
