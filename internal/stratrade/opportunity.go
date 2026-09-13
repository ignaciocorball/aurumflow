package stratrade

import "time"

type OpportunityRank struct {
	At       time.Time `json:"at"`
	Offset   string    `json:"offset"`
	Market   string    `json:"market"`
	Rank     int       `json:"rank"`
	Attention float64  `json:"attention"`
	Coverage string    `json:"coverage"`
	Tier     string    `json:"tier"`
}

type PilotRow struct {
	Market              string `json:"market"`
	AttentionCoverage   string `json:"attention_coverage"`
	MarketDataHealth    string `json:"market_data_health"`
	HistoryReadiness    string `json:"history_readiness"`
	LegacyCompatibility string `json:"legacy_compatibility"`
	BrokerSpec          string `json:"broker_spec_completeness"`
	Diversification     string `json:"diversification_value"`
	SessionAvailability string `json:"session_availability"`
	SpreadQuality       string `json:"spread_quality"`
}

func RankStability(a, b []OpportunityRank) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	pos := map[string]int{}
	for _, r := range a {
		pos[r.Market] = r.Rank
	}
	same := 0
	n := 0
	for _, r := range b {
		if prev, ok := pos[r.Market]; ok {
			n++
			if prev == r.Rank {
				same++
			}
		}
	}
	if n == 0 {
		return 0
	}
	return float64(same) / float64(n)
}

func DefaultPilotMatrix() []PilotRow {
	return []PilotRow{
		{Market: "SILVER", AttentionCoverage: "UNTESTED", MarketDataHealth: "OBSERVED", HistoryReadiness: "UNTESTED", LegacyCompatibility: "UNKNOWN", BrokerSpec: "UNTESTED", Diversification: "METALS", SessionAvailability: "OVERLAPS_GOLD", SpreadQuality: "UNTESTED"},
		{Market: "US100", AttentionCoverage: "UNTESTED", MarketDataHealth: "OBSERVED", HistoryReadiness: "UNTESTED", LegacyCompatibility: "UNKNOWN", BrokerSpec: "UNTESTED", Diversification: "US_EQUITY", SessionAvailability: "NY", SpreadQuality: "UNTESTED"},
		{Market: "OIL", AttentionCoverage: "UNTESTED", MarketDataHealth: "OBSERVED", HistoryReadiness: "UNTESTED", LegacyCompatibility: "UNKNOWN", BrokerSpec: "UNTESTED", Diversification: "ENERGY", SessionAvailability: "OVERLAPS_GOLD", SpreadQuality: "UNTESTED"},
		{Market: "US500", AttentionCoverage: "UNTESTED", MarketDataHealth: "OBSERVED", HistoryReadiness: "UNTESTED", LegacyCompatibility: "UNKNOWN", BrokerSpec: "UNTESTED", Diversification: "US_EQUITY", SessionAvailability: "NY", SpreadQuality: "UNTESTED"},
	}
}
