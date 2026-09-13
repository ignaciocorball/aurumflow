package demoprep

import (
	"sort"

	"aurumflow/internal/eligibility"
	"aurumflow/internal/instrument"
	"aurumflow/internal/opportunity"
	"aurumflow/internal/worlddomain"
)

const LegacyGoldNote = `
Legacy SignalComposer ATR buckets 8-12 / 12-18, Fib+ATR stops, and London/NY session
filters were fitted on GOLD. Compatible != validated.
`

func LegacyCompat(canonical string) string {
	switch canonical {
	case "GOLD":
		return "LEGACY_COMPATIBLE"
	case "SILVER", "OIL_CRUDE", "US100", "US500", "US30", "DE40", "UK100", "J225", "CN50":
		return "LEGACY_NEEDS_CONFIG"
	default:
		return "LEGACY_NOT_VALIDATED"
	}
}

type Row struct {
	Market           string
	Epic             string
	Discovered       bool
	Tradeable        bool
	SpecValid        bool
	MonetaryMeta     string
	RuntimeCalibrated bool
	DemoEligible     bool
	MinSize          float64
	Increment        float64
	Currency         string
	Margin           string
	StopRules        string
	MarketStatus     string
	MoneyConfidence  string
	Eligibility      worlddomain.ExecEligibility
	Legacy           string
	ReadyToCalibrate bool
}

type QueueItem struct {
	Market     string
	Tradeable  bool
	Tier       worlddomain.AttentionTier
	Relevance  int
	Eligibility worlddomain.ExecEligibility
}

func Queue(rows []Row, ranks []opportunity.Ranked) []QueueItem {
	tier := map[string]worlddomain.AttentionTier{}
	for _, r := range ranks {
		tier[r.Market] = r.Tier
	}
	var q []QueueItem
	for _, r := range rows {
		if !r.Discovered || r.Epic == "" {
			continue
		}
		rel := 0
		if r.Tradeable {
			rel += 8
		}
		if r.SpecValid {
			rel += 4
		}
		if r.Market == "GOLD" {
			rel += 2
		}
		q = append(q, QueueItem{
			Market: r.Market, Tradeable: r.Tradeable, Tier: tier[r.Market],
			Relevance: rel, Eligibility: r.Eligibility,
		})
	}
	sort.Slice(q, func(i, j int) bool {
		if q[i].Tradeable != q[j].Tradeable {
			return q[i].Tradeable
		}
		if q[i].Tier != q[j].Tier {
			return tierRank(q[i].Tier) < tierRank(q[j].Tier)
		}
		if q[i].Relevance != q[j].Relevance {
			return q[i].Relevance > q[j].Relevance
		}
		return q[i].Market < q[j].Market
	})
	return q
}

func ReadyNames(q []QueueItem) []string {
	var out []string
	for _, it := range q {
		if it.Tradeable && (it.Eligibility == worlddomain.EligSpecValid || it.Eligibility == worlddomain.EligDiscovered) {
			out = append(out, it.Market)
		}
	}
	return out
}

type Readiness struct {
	Market     string
	Data       int
	Strategy   int
	Monetary   int
	Execution  int
	Risk       int
	Operations int
	Total      int
	Notes      []string
}

func ScoreReadiness(r Row, cov float64, e eligibility.Entry) Readiness {
	out := Readiness{Market: r.Market}
	out.Data = clamp(int(cov))
	switch LegacyCompat(r.Market) {
	case "LEGACY_COMPATIBLE":
		out.Strategy = 70
	case "LEGACY_NEEDS_CONFIG":
		out.Strategy = 35
	default:
		out.Strategy = 10
	}
	if r.RuntimeCalibrated {
		out.Monetary = 90
	} else if r.SpecValid {
		out.Monetary = 55
	} else if r.Discovered {
		out.Monetary = 25
	}
	switch e.Status {
	case worlddomain.EligDemo:
		out.Execution = 90
	case worlddomain.EligCalibrated:
		out.Execution = 70
	case worlddomain.EligSpecValid:
		out.Execution = 45
	case worlddomain.EligDiscovered:
		out.Execution = 25
	default:
		out.Execution = 5
	}
	if r.MoneyConfidence == "UNKNOWN" || r.MonetaryMeta == "UNKNOWN" {
		out.Risk = 0
		out.Notes = append(out.Notes, "unknown monetary risk blocks new order")
	} else {
		out.Risk = 40
	}
	if r.Tradeable {
		out.Operations = 70
	} else {
		out.Operations = 20
		out.Notes = append(out.Notes, "broker status not TRADEABLE")
	}
	out.Total = (out.Data + out.Strategy + out.Monetary + out.Execution + out.Risk + out.Operations) / 6
	return out
}

func tierRank(t worlddomain.AttentionTier) int {
	switch t {
	case worlddomain.TierA:
		return 0
	case worlddomain.TierB:
		return 1
	case worlddomain.TierC:
		return 2
	default:
		return 3
	}
}

func clamp(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func UniverseRows(maps []instrument.Mapping) []Row {
	idx := map[string]instrument.Mapping{}
	for _, m := range maps {
		idx[m.Canonical] = m
	}
	var out []Row
	for _, id := range instrument.PrepareUniverse() {
		m := idx[id]
		r := Row{
			Market: id, Epic: m.Epic, Discovered: m.Epic != "",
			Currency: m.Currency, MonetaryMeta: "UNKNOWN", MoneyConfidence: "UNKNOWN",
			Eligibility: worlddomain.EligAnalysis, Legacy: LegacyCompat(id),
			StopRules: "UNKNOWN", Margin: "UNKNOWN",
		}
		if m.Unresolved || (id == "US500" && m.Epic == "") {
			r.Eligibility = worlddomain.EligAnalysis
			r.MonetaryMeta = "UNRESOLVED"
		} else if m.Epic != "" {
			r.Eligibility = worlddomain.EligDiscovered
		}
		out = append(out, r)
	}
	return out
}
