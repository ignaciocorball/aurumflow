package instrument

import (
	"encoding/json"
	"os"
	"path/filepath"

	"aurumflow/internal/worlddomain"
)

type CanonicalMarket struct {
	ID          string
	Region      worlddomain.Region
	AssetClass  worlddomain.AssetClass
	Underlying  string
	Execution   VenueInstrument
	Sensor      VenueInstrument
	Relation    string
	Eligibility worlddomain.ExecEligibility
}

type Mapping struct {
	Canonical   string `json:"canonical"`
	Epic        string `json:"capital_epic"`
	Name        string `json:"name"`
	Region      string `json:"region"`
	Asset       string `json:"asset_class"`
	Currency    string `json:"currency,omitempty"`
	BrokerType  string `json:"broker_instrument_type,omitempty"`
	Underlying  string `json:"underlying,omitempty"`
	Unresolved  bool   `json:"unresolved,omitempty"`
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

func DefaultMarkets() []CanonicalMarket {
	r := DefaultRegistry()
	gold, _ := r.ExecutionFor("GOLD")
	btc, _ := r.ExecutionFor("BITCOIN")
	us100, _ := r.ExecutionFor("NASDAQ100")
	return []CanonicalMarket{
		{ID: "GOLD", Region: worlddomain.RegionGlobal, AssetClass: worlddomain.AssetPrecious, Underlying: "XAU", Execution: gold, Relation: RelNone, Eligibility: worlddomain.EligAnalysis},
		{ID: "BTC", Region: worlddomain.RegionGlobal, AssetClass: worlddomain.AssetCrypto, Underlying: "BTC", Execution: btc, Relation: RelProxy, Eligibility: worlddomain.EligAnalysis},
		{ID: "US100", Region: worlddomain.RegionUS, AssetClass: worlddomain.AssetEquities, Underlying: "NDX", Execution: us100, Relation: RelProxy, Eligibility: worlddomain.EligAnalysis},
	}
}

func TapeMarkets() []string {
	return []string{"US100", "US500", "US30", "GOLD", "SILVER", "OIL_CRUDE", "DE40", "UK100", "J225", "CN50", "BTC"}
}

func PrepareUniverse() []string {
	return []string{"GOLD", "SILVER", "OIL_CRUDE", "US100", "US500", "US30", "DE40", "UK100", "J225", "CN50"}
}
