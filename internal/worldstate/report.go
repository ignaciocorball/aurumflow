package worldstate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"aurumflow/internal/worlddomain"
)

type RankView struct {
	Rank      int
	Market    string
	Score     float64
	Coverage  float64
	Tier      string
	Why       []string
	Risks     []string
	Freshness string
	State     MarketState
}

func WriteJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func WriteMD(path, body string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(body), 0o644)
}

func CurrentMarkdown(ws WorldState, ranks []RankView) string {
	var b strings.Builder
	b.WriteString("# CURRENT WORLD STATE\n\n")
	b.WriteString("AsOf: " + ws.AsOf.UTC().Format(time.RFC3339) + "\n")
	b.WriteString("Valid: " + ws.Valid + "\n")
	b.WriteString("Origins: " + strings.Join(ws.Origins, ",") + "\n")
	b.WriteString("Session: " + string(ws.Session) + "\n")
	b.WriteString("Global liquidity: " + string(ws.Liquidity.Class) + " (" + ws.Liquidity.Label + ")\n")
	b.WriteString("Risk regime: " + string(ws.Risk) + "\n")
	b.WriteString("USD: " + ws.USD.USD + "\n")
	b.WriteString("Rates: US=")
	if ws.Rates.USPresent {
		b.WriteString(f(ws.Rates.USPolicy))
	} else {
		b.WriteString(worlddomain.MissingLabel())
	}
	b.WriteString(" EU=")
	if ws.Rates.EUPresent {
		b.WriteString(f(ws.Rates.EUPolicy))
	} else {
		b.WriteString(worlddomain.MissingLabel())
	}
	b.WriteString("\n\n## Regions\n")
	keys := []worlddomain.Region{worlddomain.RegionUS, worlddomain.RegionEurope, worlddomain.RegionJapan, worlddomain.RegionChinaHK}
	for _, k := range keys {
		r := ws.Regions[k]
		b.WriteString("- " + string(k) + " equity=" + nz(r.Equity) + " health=" + string(r.Health) + "\n")
	}
	b.WriteString("\n## Markets\n")
	var mk []string
	for id := range ws.Markets {
		mk = append(mk, id)
	}
	sort.Strings(mk)
	for _, id := range mk {
		m := ws.Markets[id]
		b.WriteString("- " + id + " flow=" + nz(m.CapitalFlowContext) + " pos=" + nz(m.Positioning) + " q=" + string(m.DataQuality) + "\n")
	}
	b.WriteString("\n## Attention (not a trade list)\n")
	for _, r := range ranks {
		b.WriteString("- #" + itoa(r.Rank) + " " + r.Market + " score=" + f(r.Score) + " tier=" + r.Tier + "\n")
	}
	return b.String()
}

func OpportunityMarkdown(ranks []RankView) string {
	var b strings.Builder
	b.WriteString("# OPPORTUNITY SURFACE\n\nATTENTION_SCORE_V1 — where to look, not what to buy.\n\n")
	for _, r := range ranks {
		b.WriteString("## " + itoa(r.Rank) + ". " + r.Market + "\n")
		b.WriteString("Score: " + f(r.Score) + "\nState: " + r.State.PriceTrend + "\n")
		b.WriteString("Why: " + strings.Join(r.Why, "; ") + "\n")
		b.WriteString("Risk: " + strings.Join(r.Risks, "; ") + "\n")
		b.WriteString("Freshness: " + string(r.Freshness) + "\n")
		b.WriteString("Micro available: " + yn(r.State.MicroAvailable) + "\n\n")
	}
	return b.String()
}

func Capital12Markdown(ws WorldState, snaps int, coverage string) string {
	var b strings.Builder
	b.WriteString("# GLOBAL CAPITAL 12M\n\n")
	b.WriteString("Snapshots: " + itoa(snaps) + "\nCoverage: " + coverage + "\n")
	b.WriteString("No-lookahead: WorldState.At filters AvailableAt.\n")
	b.WriteString("Revisions: period + retrieved_at + available_at stored on ContextObservation.\n\n")
	b.WriteString("Where did capital move (observed proxies only):\n")
	for _, f := range ws.Flows {
		b.WriteString("- " + string(f.Region) + " " + string(f.AssetClass) + " " + string(f.Direction) + " strength=" + fstr(f.Strength) + " " + strings.Join(f.Evidence, ",") + "\n")
	}
	b.WriteString("\nGlobal liquidity: " + string(ws.Liquidity.Class) + "\n")
	b.WriteString("Risk regime: " + string(ws.Risk) + "\n")
	return b.String()
}

func nz(s string) string {
	if s == "" {
		return worlddomain.MissingLabel()
	}
	return s
}

func yn(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func f(v float64) string {
	return strings.TrimRight(strings.TrimRight(sprintf(v), "0"), ".")
}

func fstr(v float64) string { return f(v) }

func sprintf(v float64) string {
	neg := v < 0
	if neg {
		v = -v
	}
	ip := int(v)
	fp := int((v-float64(ip))*100 + 0.5)
	s := itoa(ip) + "." + pad2(fp)
	if neg {
		return "-" + s
	}
	return s
}

func pad2(n int) string {
	if n < 10 {
		return "0" + itoa(n)
	}
	return itoa(n)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [16]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
