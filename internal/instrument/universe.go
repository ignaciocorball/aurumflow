package instrument

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"aurumflow/internal/worlddomain"
)

type CanonicalMarket struct {
	ID           string
	Region       worlddomain.Region
	AssetClass   worlddomain.AssetClass
	Underlying   string
	Execution    VenueInstrument
	Sensor       VenueInstrument
	Relation     string
	Eligibility  worlddomain.ExecEligibility
}

type Mapping struct {
	Canonical string `json:"canonical"`
	Epic      string `json:"capital_epic"`
	Name      string `json:"name"`
	Region    string `json:"region"`
	Asset     string `json:"asset_class"`
}

func SearchTerms() []string {
	return []string{"Gold", "Silver", "Oil", "Crude", "Nasdaq", "S&P", "Dow", "Germany", "UK", "Japan", "Hong Kong", "China"}
}

func PersistMappings(path string, ms []Mapping) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(ms, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func LoadMappings(path string) ([]Mapping, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var ms []Mapping
	if err := json.Unmarshal(raw, &ms); err != nil {
		return nil, err
	}
	return ms, nil
}

func GuessCanonical(name, epic string) (string, worlddomain.Region, worlddomain.AssetClass) {
	s := strings.ToUpper(name + " " + epic)
	switch {
	case strings.Contains(s, "GOLD") && !strings.Contains(s, "GOLDMAN"):
		return "GOLD", worlddomain.RegionGlobal, worlddomain.AssetPrecious
	case strings.Contains(s, "SILVER"):
		return "SILVER", worlddomain.RegionGlobal, worlddomain.AssetPrecious
	case strings.Contains(s, "OIL") || strings.Contains(s, "CRUDE") || strings.Contains(s, "WTI") || strings.Contains(s, "BRENT"):
		return "OIL", worlddomain.RegionGlobal, worlddomain.AssetEnergy
	case strings.Contains(s, "US100") || strings.Contains(s, "NASDAQ") || strings.Contains(s, "NDX"):
		return "US100", worlddomain.RegionUS, worlddomain.AssetEquities
	case strings.Contains(s, "US500") || strings.Contains(s, "S&P") || strings.Contains(s, "SPX"):
		return "US500", worlddomain.RegionUS, worlddomain.AssetEquities
	case strings.Contains(s, "US30") || strings.Contains(s, "DOW"):
		return "US30", worlddomain.RegionUS, worlddomain.AssetEquities
	case strings.Contains(s, "GER") || strings.Contains(s, "DAX") || strings.Contains(s, "GERMANY"):
		return "EUROPE", worlddomain.RegionEurope, worlddomain.AssetEquities
	case strings.Contains(s, "UK100") || strings.Contains(s, "FTSE"):
		return "UK", worlddomain.RegionEurope, worlddomain.AssetEquities
	case strings.Contains(s, "JAPAN") || strings.Contains(s, "JP225") || strings.Contains(s, "NIKKEI"):
		return "JAPAN", worlddomain.RegionJapan, worlddomain.AssetEquities
	case strings.Contains(s, "HONG") || strings.Contains(s, "HK50") || strings.Contains(s, "HANG"):
		return "HK", worlddomain.RegionChinaHK, worlddomain.AssetEquities
	case strings.Contains(s, "CHINA") || strings.Contains(s, "HSCE"):
		return "CHINA", worlddomain.RegionChinaHK, worlddomain.AssetEquities
	case strings.Contains(s, "BTC") || strings.Contains(s, "BITCOIN"):
		return "BTC", worlddomain.RegionGlobal, worlddomain.AssetCrypto
	default:
		return "", "", ""
	}
}

func DefaultMarkets() []CanonicalMarket {
	r := DefaultRegistry()
	gold, _ := r.ExecutionFor("GOLD")
	btc, _ := r.ExecutionFor("BITCOIN")
	us100, _ := r.ExecutionFor("NASDAQ100")
	return []CanonicalMarket{
		{ID: "GOLD", Region: worlddomain.RegionGlobal, AssetClass: worlddomain.AssetPrecious, Underlying: "XAU", Execution: gold, Relation: RelNone, Eligibility: worlddomain.EligNotCalibrated},
		{ID: "BTC", Region: worlddomain.RegionGlobal, AssetClass: worlddomain.AssetCrypto, Underlying: "BTC", Execution: btc, Relation: RelProxy, Eligibility: worlddomain.EligAnalysis},
		{ID: "US100", Region: worlddomain.RegionUS, AssetClass: worlddomain.AssetEquities, Underlying: "NDX", Execution: us100, Relation: RelProxy, Eligibility: worlddomain.EligAnalysis},
	}
}
