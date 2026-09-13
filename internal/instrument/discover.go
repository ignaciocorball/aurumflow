package instrument

import (
	"strings"

	"aurumflow/internal/worlddomain"
)

type SearchHit struct {
	Epic, Name, Status, InstrumentType, Currency, Country, Region string
}

func MapHits(hits []SearchHit) []Mapping {
	var out []Mapping
	seen := map[string]bool{}
	var us500 []CatalogHit
	for _, h := range hits {
		if strings.TrimSpace(h.Epic) == "" {
			continue
		}
		if id, reg, ac, ok := ExactEpic(h.Epic); ok {
			if seen[id] {
				continue
			}
			seen[id] = true
			out = append(out, Mapping{
				Canonical: id, Epic: h.Epic, Name: h.Name,
				Region: string(reg), Asset: string(ac),
				Currency: h.Currency, BrokerType: h.InstrumentType,
				Underlying: underlyingOf(id),
			})
			continue
		}
	}
	if !seen["US500"] {
		for _, h := range hits {
			if strings.TrimSpace(h.Epic) == "" {
				continue
			}
			if _, _, _, ok := ExactEpic(h.Epic); ok {
				continue
			}
			if looksUS500(strings.ToUpper(h.Name + " " + h.Epic)) {
				us500 = append(us500, CatalogHit{
					Epic: h.Epic, Name: h.Name, InstrumentType: h.InstrumentType,
					Currency: h.Currency, Country: h.Country, Region: h.Region, Status: h.Status,
				})
			}
		}
		if m, ok := ResolveUS500(us500); ok {
			out = append(out, m)
			seen["US500"] = true
		} else if len(us500) > 0 {
			m.Unresolved = true
			out = append(out, m)
		}
	}
	for _, h := range hits {
		if strings.TrimSpace(h.Epic) == "" {
			continue
		}
		if _, _, _, ok := ExactEpic(h.Epic); ok {
			continue
		}
		id, reg, ac := GuessCanonical(h.Name, h.Epic)
		if id == "" || id == "US500" {
			continue
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, Mapping{
			Canonical: id, Epic: h.Epic, Name: h.Name,
			Region: string(reg), Asset: string(ac),
			Currency: h.Currency, BrokerType: h.InstrumentType,
			Underlying: underlyingOf(id),
		})
	}
	return out
}

func underlyingOf(id string) string {
	switch id {
	case "GOLD":
		return "XAU"
	case "SILVER":
		return "XAG"
	case "OIL_CRUDE":
		return "WTI"
	case "US100":
		return "NDX"
	case "US500":
		return "SPX"
	case "US30":
		return "DJI"
	case "DE40":
		return "DAX"
	case "UK100":
		return "UKX"
	case "J225":
		return "NKY"
	case "CN50":
		return "XIN9"
	case "BTC":
		return "BTC"
	default:
		return id
	}
}

func EligibilityFor(id string) worlddomain.ExecEligibility {
	return worlddomain.EligAnalysis
}
