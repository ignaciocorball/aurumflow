package mktval

import (
	"fmt"
	"strings"
)

func MatrixMarkdown(rows []Row) string {
	var b strings.Builder
	b.WriteString("# MARKET VALIDATION MATRIX\n\n")
	b.WriteString("| Market | Data | History | Mechanics | Legacy | Norm V0 | Disc | Val | Hold | Sig | Exp | PF | MaxDD | Costs | Subs | Broker | Monetary | Shadow | DEMO | Trust | Status |\n")
	b.WriteString("|---|---|---|---|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|---|---|---|---|---|\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %d | %d | %d | %d | %.3f | %.2f | %.2f | %.3f | %d | %s | %s | %s | %s | %s | %s |\n",
			r.Market, r.Data, r.History, r.Mechanics, r.LegacyCompat, r.NormCompat,
			r.DiscoveryN, r.ValidationN, r.HoldoutN, r.Signals, r.Expectancy, r.ProfitFactor, r.MaxDD, r.CostDrag, r.PositiveSubs,
			r.BrokerSpec, r.Monetary, r.Shadow, r.DemoElig, r.Trust, r.Status)
	}
	b.WriteString("\nPromotion uses holdout expectancy after NORMAL costs. Attention/Salience do not promote.\n")
	return b.String()
}

func RankingMarkdown(rows []Row) string {
	var b strings.Builder
	b.WriteString("# MARKET RANKING FOR PROMOTION\n\n")
	b.WriteString("| Market | Holdout N | Net exp | PF | MaxDD | +subs | cost drag | data | exec | status |\n")
	b.WriteString("|---|---:|---:|---:|---:|---:|---:|---|---|---|\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "| %s | %d | %.3f | %.2f | %.2f | %d | %.3f | %s | %s | %s |\n",
			r.Market, r.HoldoutN, r.Expectancy, r.ProfitFactor, r.MaxDD, r.PositiveSubs, r.CostDrag, r.Data, r.BrokerSpec, r.Status)
	}
	return b.String()
}
