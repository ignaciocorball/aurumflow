package portfoliorisk

import "strings"

const (
	GroupUSEquity   = "US_EQUITY_INDEX"
	GroupPrecious   = "PRECIOUS_METALS"
	GroupEnergy     = "ENERGY"
	GroupEurope     = "EUROPE_EQUITY"
	GroupAsia       = "ASIA_EQUITY"
	GroupCrypto     = "CRYPTO"
)

type Caps struct {
	MaxOpenPositions     int
	MaxCorrelated        int
	MaxAggregateRisk     float64
	MaxRiskPerRegion     float64
	MaxRiskPerAssetClass float64
}

func ConservativeCaps() Caps {
	return Caps{
		MaxOpenPositions:     3,
		MaxCorrelated:        1,
		MaxAggregateRisk:     300,
		MaxRiskPerRegion:     150,
		MaxRiskPerAssetClass: 150,
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
	Pass             bool
	Blocks           []string
	GrossRisk        float64
	ByAsset          map[string]float64
	ByRegion         map[string]float64
	ByGroup          map[string]float64
	OpenPositions    int
	CorrelatedOpen   int
}

type Manager struct {
	Caps Caps
}

func New() *Manager { return &Manager{Caps: ConservativeCaps()} }

func GroupOf(canonical string) string {
	switch strings.ToUpper(canonical) {
	case "US100", "US500", "US30":
		return GroupUSEquity
	case "GOLD", "SILVER":
		return GroupPrecious
	case "OIL", "CRUDE", "CRUDE_OIL":
		return GroupEnergy
	case "EUROPE", "UK", "GERMANY":
		return GroupEurope
	case "JAPAN", "HK", "CHINA":
		return GroupAsia
	case "BTC", "ETH", "BITCOIN":
		return GroupCrypto
	default:
		return "OTHER"
	}
}

func (m *Manager) Evaluate(open []Position, candidate Position) Result {
	r := Result{Pass: true, ByAsset: map[string]float64{}, ByRegion: map[string]float64{}, ByGroup: map[string]float64{}}
	add := func(p Position) {
		r.OpenPositions++
		if !p.RiskKnown {
			r.Pass = false
			r.Blocks = append(r.Blocks, "unknown monetary risk blocks new order")
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
	return r
}
