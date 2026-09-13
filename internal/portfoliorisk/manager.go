package portfoliorisk

import (
	"sort"
	"strings"
)

const (
	GroupUSEquity = "US_EQUITY_INDEX"
	GroupPrecious = "PRECIOUS_METALS"
	GroupEnergy   = "ENERGY"
	GroupEurope   = "EUROPE_EQUITY"
	GroupAsia     = "ASIA_EQUITY"
	GroupCrypto   = "CRYPTO"

	// RiskUnit is account-currency risk: expected loss if each position hits its
	// configured stop, measured in the DEMO account currency (typically USD).
	// Caps below are therefore account-currency risk budgets, not dimensionless counts
	// and not notional exposure.
	RiskUnit = "account_currency_risk"
)

type Caps struct {
	MaxOpenPositions     int
	MaxCorrelated        int
	MaxAggregateRisk     float64
	MaxRiskPerRegion     float64
	MaxRiskPerAssetClass float64
	Unit                 string
}

func ConservativeCaps() Caps {
	return Caps{
		MaxOpenPositions:     3,
		MaxCorrelated:        1,
		MaxAggregateRisk:     300,
		MaxRiskPerRegion:     150,
		MaxRiskPerAssetClass: 150,
		Unit:                 RiskUnit,
	}
}

type Position struct {
	Canonical string
	Region    string
	Asset     string
	RiskMoney float64
	RiskKnown bool
}

type Result struct {
	Pass              bool
	Blocks            []string
	GrossRisk         float64
	ByAsset           map[string]float64
	ByRegion          map[string]float64
	ByGroup           map[string]float64
	ByMarket          map[string]float64
	OpenPositions     int
	CorrelatedOpen    int
	RemainingAggregate float64
	RemainingRegion    map[string]float64
	RemainingAsset     map[string]float64
	Unit               string
}

type Manager struct {
	Caps Caps
}

func New() *Manager { return &Manager{Caps: ConservativeCaps()} }

func GroupOf(canonical string) string {
	switch strings.ToUpper(canonical) {
	case "US100", "US500", "US30", "AAPL", "MSFT", "NVDA", "META", "AMZN":
		return GroupUSEquity
	case "GOLD", "SILVER":
		return GroupPrecious
	case "OIL", "OIL_CRUDE", "CRUDE", "CRUDE_OIL":
		return GroupEnergy
	case "DE40", "UK100", "EUROPE", "UK", "GERMANY":
		return GroupEurope
	case "J225", "CN50", "JAPAN", "HK", "CHINA", "CHINA_HK":
		return GroupAsia
	case "BTC", "ETH", "BITCOIN":
		return GroupCrypto
	default:
		return "OTHER"
	}
}

func UniqueSorted(xs []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, x := range xs {
		x = strings.TrimSpace(x)
		if x == "" || seen[x] {
			continue
		}
		seen[x] = true
		out = append(out, x)
	}
	sort.Strings(out)
	return out
}

func (m *Manager) Evaluate(open []Position, candidate Position) Result {
	r := Result{
		Pass: true, Unit: m.Caps.Unit,
		ByAsset: map[string]float64{}, ByRegion: map[string]float64{},
		ByGroup: map[string]float64{}, ByMarket: map[string]float64{},
		RemainingRegion: map[string]float64{}, RemainingAsset: map[string]float64{},
	}
	unknown := false
	add := func(p Position) {
		r.OpenPositions++
		r.ByMarket[p.Canonical] += p.RiskMoney
		if !p.RiskKnown {
			unknown = true
			return
		}
		r.GrossRisk += p.RiskMoney
		r.ByAsset[p.Asset] += p.RiskMoney
		r.ByRegion[p.Region] += p.RiskMoney
		r.ByGroup[GroupOf(p.Canonical)] += p.RiskMoney
	}
	for _, p := range open {
		add(p)
	}
	if unknown {
		r.Pass = false
		r.Blocks = append(r.Blocks, "unknown monetary risk blocks new order")
	}
	if candidate.Canonical != "" {
		if !candidate.RiskKnown {
			r.Pass = false
			r.Blocks = append(r.Blocks, "unknown monetary risk blocks new order")
		} else {
			g := GroupOf(candidate.Canonical)
			same := 0
			for _, p := range open {
				if GroupOf(p.Canonical) == g {
					same++
				}
			}
			r.CorrelatedOpen = same
			if same >= m.Caps.MaxCorrelated {
				r.Pass = false
				r.Blocks = append(r.Blocks, "portfolio "+g+" risk cap")
			}
			nextGross := r.GrossRisk + candidate.RiskMoney
			if nextGross > m.Caps.MaxAggregateRisk {
				r.Pass = false
				r.Blocks = append(r.Blocks, "max aggregate risk")
			}
			if r.ByRegion[candidate.Region]+candidate.RiskMoney > m.Caps.MaxRiskPerRegion {
				r.Pass = false
				r.Blocks = append(r.Blocks, "max risk per region")
			}
			if r.ByAsset[candidate.Asset]+candidate.RiskMoney > m.Caps.MaxRiskPerAssetClass {
				r.Pass = false
				r.Blocks = append(r.Blocks, "max risk per asset class")
			}
			if r.OpenPositions+1 > m.Caps.MaxOpenPositions {
				r.Pass = false
				r.Blocks = append(r.Blocks, "max total open positions")
			}
		}
	}
	r.Blocks = UniqueSorted(r.Blocks)
	r.RemainingAggregate = m.Caps.MaxAggregateRisk - r.GrossRisk
	for k, v := range r.ByRegion {
		r.RemainingRegion[k] = m.Caps.MaxRiskPerRegion - v
	}
	for k, v := range r.ByAsset {
		r.RemainingAsset[k] = m.Caps.MaxRiskPerAssetClass - v
	}
	return r
}
