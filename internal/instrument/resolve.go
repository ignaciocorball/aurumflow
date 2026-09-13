package instrument

import (
	"strings"

	"aurumflow/internal/worlddomain"
)

type CatalogHit struct {
	Epic, Name, InstrumentType, Currency, Country, Region, Status string
}

func SearchTerms() []string {
	return []string{
		"Gold", "Silver", "Oil", "Crude",
		"US100", "Nasdaq", "US30", "Dow",
		"S&P 500", "US500",
		"DE40", "DAX", "UK100", "FTSE",
		"J225", "Nikkei", "CN50",
		"AAPL", "MSFT", "NVDA", "META", "AMZN",
	}
}

func ExactEpic(epic string) (string, worlddomain.Region, worlddomain.AssetClass, bool) {
	switch strings.ToUpper(strings.TrimSpace(epic)) {
	case "GOLD":
		return "GOLD", worlddomain.RegionGlobal, worlddomain.AssetPrecious, true
	case "SILVER":
		return "SILVER", worlddomain.RegionGlobal, worlddomain.AssetPrecious, true
	case "OIL_CRUDE":
		return "OIL_CRUDE", worlddomain.RegionGlobal, worlddomain.AssetEnergy, true
	case "US100":
		return "US100", worlddomain.RegionUS, worlddomain.AssetEquities, true
	case "US500":
		return "US500", worlddomain.RegionUS, worlddomain.AssetEquities, true
	case "US30":
		return "US30", worlddomain.RegionUS, worlddomain.AssetEquities, true
	case "DE40":
		return "DE40", worlddomain.RegionEurope, worlddomain.AssetEquities, true
	case "UK100":
		return "UK100", worlddomain.RegionEurope, worlddomain.AssetEquities, true
	case "J225":
		return "J225", worlddomain.RegionJapan, worlddomain.AssetEquities, true
	case "CN50":
		return "CN50", worlddomain.RegionChinaHK, worlddomain.AssetEquities, true
	default:
		return "", "", "", false
	}
}

func GuessCanonical(name, epic string) (string, worlddomain.Region, worlddomain.AssetClass) {
	if id, reg, ac, ok := ExactEpic(epic); ok {
		return id, reg, ac
	}
	s := strings.ToUpper(name + " " + epic)
	switch {
	case strings.Contains(s, "GOLD") && !strings.Contains(s, "GOLDMAN"):
		return "GOLD", worlddomain.RegionGlobal, worlddomain.AssetPrecious
	case strings.Contains(s, "SILVER"):
		return "SILVER", worlddomain.RegionGlobal, worlddomain.AssetPrecious
	case strings.Contains(s, "OIL") || strings.Contains(s, "CRUDE") || strings.Contains(s, "WTI") || strings.Contains(s, "BRENT"):
		return "OIL_CRUDE", worlddomain.RegionGlobal, worlddomain.AssetEnergy
	case strings.Contains(s, "US100") || strings.Contains(s, "NASDAQ 100") || strings.Contains(s, "NASDAQ100") || strings.Contains(s, "NDX"):
		return "US100", worlddomain.RegionUS, worlddomain.AssetEquities
	case exactUS500Token(s):
		return "US500", worlddomain.RegionUS, worlddomain.AssetEquities
	case strings.Contains(s, "US30") || strings.Contains(s, "DOW JONES") || strings.Contains(s, "DJIA") || strings.Contains(s, "WALL STREET 30"):
		return "US30", worlddomain.RegionUS, worlddomain.AssetEquities
	case strings.Contains(s, "DE40") || strings.Contains(s, "DAX") || (strings.Contains(s, "GERMANY") && strings.Contains(s, "40")):
		return "DE40", worlddomain.RegionEurope, worlddomain.AssetEquities
	case strings.Contains(s, "UK100") || strings.Contains(s, "FTSE"):
		return "UK100", worlddomain.RegionEurope, worlddomain.AssetEquities
	case strings.Contains(s, "J225") || strings.Contains(s, "JP225") || strings.Contains(s, "NIKKEI"):
		return "J225", worlddomain.RegionJapan, worlddomain.AssetEquities
	case strings.Contains(s, "CN50") || strings.Contains(s, "HSCE") || strings.Contains(s, "CHINA 50"):
		return "CN50", worlddomain.RegionChinaHK, worlddomain.AssetEquities
	case strings.Contains(s, "HONG") || strings.Contains(s, "HK50") || strings.Contains(s, "HANG"):
		return "CHINA_HK", worlddomain.RegionChinaHK, worlddomain.AssetEquities
	case strings.Contains(s, "BTC") || strings.Contains(s, "BITCOIN"):
		return "BTC", worlddomain.RegionGlobal, worlddomain.AssetCrypto
	case strings.Contains(s, "AAPL"):
		return "AAPL", worlddomain.RegionUS, worlddomain.AssetEquities
	case strings.Contains(s, "MSFT"):
		return "MSFT", worlddomain.RegionUS, worlddomain.AssetEquities
	case strings.Contains(s, "NVDA"):
		return "NVDA", worlddomain.RegionUS, worlddomain.AssetEquities
	case strings.Contains(s, "META"):
		return "META", worlddomain.RegionUS, worlddomain.AssetEquities
	case strings.Contains(s, "AMZN"):
		return "AMZN", worlddomain.RegionUS, worlddomain.AssetEquities
	default:
		return "", "", ""
	}
}

func exactUS500Token(s string) bool {
	return strings.Contains(s, "US500") || strings.Contains(s, "SPX")
}

// ResolveUS500 requires a unique broker identity. It does not pick the first S&P hit.
func ResolveUS500(hits []CatalogHit) (Mapping, bool) {
	var cand []CatalogHit
	seenEpic := map[string]bool{}
	for _, h := range hits {
		if strings.TrimSpace(h.Epic) == "" {
			continue
		}
		if seenEpic[strings.ToUpper(h.Epic)] {
			continue
		}
		name := strings.ToUpper(h.Name + " " + h.Epic)
		if !looksUS500(name) {
			continue
		}
		if looksNotUS500(name) {
			continue
		}
		if h.InstrumentType != "" && !us500TypeOK(h.InstrumentType) {
			continue
		}
		if h.Currency != "" && !strings.EqualFold(h.Currency, "USD") {
			continue
		}
		if region := strings.ToUpper(h.Country + " " + h.Region); region != "" && !us500RegionOK(region) {
			continue
		}
		seenEpic[strings.ToUpper(h.Epic)] = true
		cand = append(cand, h)
	}
	if len(cand) != 1 {
		return Mapping{Canonical: "US500", Region: string(worlddomain.RegionUS), Asset: string(worlddomain.AssetEquities)}, false
	}
	h := cand[0]
	return Mapping{
		Canonical: "US500", Epic: h.Epic, Name: h.Name,
		Region: string(worlddomain.RegionUS), Asset: string(worlddomain.AssetEquities),
		Currency: "USD", BrokerType: h.InstrumentType, Underlying: "SPX",
	}, true
}

func looksUS500(s string) bool {
	return strings.Contains(s, "US500") || strings.Contains(s, "S&P 500") || strings.Contains(s, "S&P500") ||
		strings.Contains(s, "SPX") || (strings.Contains(s, "S&P") && strings.Contains(s, "500"))
}

func looksNotUS500(s string) bool {
	return strings.Contains(s, "200") || strings.Contains(s, "400") || strings.Contains(s, "600") ||
		strings.Contains(s, "MICRO") || strings.Contains(s, "MINI") || strings.Contains(s, "DIVIDEND")
}

func us500TypeOK(t string) bool {
	u := strings.ToUpper(t)
	return strings.Contains(u, "INDICES") || strings.Contains(u, "INDEX") || strings.Contains(u, "EQUIT")
}

func us500RegionOK(s string) bool {
	return strings.Contains(s, "US") || strings.Contains(s, "UNITED") || strings.Contains(s, "AMERICA")
}
