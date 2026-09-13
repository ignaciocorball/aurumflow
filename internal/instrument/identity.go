package instrument

import "strings"

const (
	RelIdentical = "IDENTICAL"
	RelProxy     = "proxy"
	RelCorrelated = "correlated_underlying_proxy"
	RelCorrelatedProxy = "CORRELATED_PROXY"
	RelNone      = "none"
)

// CanonicalAsset is the economic thing we reason about, not a venue symbol.
type CanonicalAsset struct {
	ID   string
	Name string
}

// VenueInstrument is a specific listed product at a venue.
type VenueInstrument struct {
	Venue      string
	Symbol     string
	Class      string // cfd, future, spot, option
	Role       string // sensor, execution
	AssetID    string
	Relation   string
}

type Registry struct {
	Assets []CanonicalAsset
	Venues []VenueInstrument
}

func DefaultRegistry() Registry {
	return Registry{
		Assets: []CanonicalAsset{
			{ID: "GOLD", Name: "Gold"},
			{ID: "BITCOIN", Name: "Bitcoin"},
			{ID: "NASDAQ100", Name: "Nasdaq-100"},
		},
		Venues: []VenueInstrument{
			{Venue: "capital.com", Symbol: "GOLD", Class: "cfd", Role: "execution", AssetID: "GOLD", Relation: RelNone},
			{Venue: "cme", Symbol: "GC", Class: "future", Role: "sensor", AssetID: "GOLD", Relation: RelCorrelated},
			{Venue: "capital.com", Symbol: "BTCUSD", Class: "cfd", Role: "execution", AssetID: "BITCOIN", Relation: RelProxy},
			{Venue: "binance_usdm", Symbol: "BTCUSDT", Class: "future", Role: "sensor", AssetID: "BITCOIN", Relation: RelProxy},
			{Venue: "okx", Symbol: "BTC-USDT-SWAP", Class: "future", Role: "sensor", AssetID: "BITCOIN", Relation: RelCorrelatedProxy},
			{Venue: "capital.com", Symbol: "US100", Class: "cfd", Role: "execution", AssetID: "NASDAQ100", Relation: RelProxy},
			{Venue: "cme", Symbol: "NQ", Class: "future", Role: "sensor", AssetID: "NASDAQ100", Relation: RelCorrelated},
		},
	}
}

func (r Registry) ExecutionFor(assetID string) (VenueInstrument, bool) {
	for _, v := range r.Venues {
		if v.AssetID == assetID && v.Role == "execution" {
			return v, true
		}
	}
	return VenueInstrument{}, false
}

func (r Registry) SensorsFor(assetID string) []VenueInstrument {
	var out []VenueInstrument
	for _, v := range r.Venues {
		if v.AssetID == assetID && v.Role == "sensor" {
			out = append(out, v)
		}
	}
	return out
}

func SameInstrument(a, b VenueInstrument) bool {
	return strings.EqualFold(a.Venue, b.Venue) && strings.EqualFold(a.Symbol, b.Symbol) && strings.EqualFold(a.Class, b.Class)
}
