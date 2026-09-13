package microcap

import "strings"

// Capability is sensing availability only. It is not a bullish, bearish, or trade signal.
type Capability struct {
	Market         string
	Trades         bool
	IncrementalL2  bool
	Pressure       bool
	Exhaustion     bool
	Absorption     bool
	ProviderHealth string
}

type Registry struct {
	items map[string]Capability
}

func New() *Registry { return &Registry{items: map[string]Capability{}} }

func (r *Registry) Set(c Capability) {
	if r.items == nil {
		r.items = map[string]Capability{}
	}
	c.Market = strings.ToUpper(strings.TrimSpace(c.Market))
	r.items[c.Market] = c
}

func (r *Registry) Get(market string) Capability {
	if r == nil || r.items == nil {
		return Capability{Market: strings.ToUpper(market), ProviderHealth: "UNKNOWN"}
	}
	c, ok := r.items[strings.ToUpper(strings.TrimSpace(market))]
	if !ok {
		return Capability{Market: strings.ToUpper(market), ProviderHealth: "UNKNOWN"}
	}
	return c
}

func (c Capability) Available() bool {
	return c.Trades || c.IncrementalL2 || c.Pressure || c.Exhaustion || c.Absorption
}

func (c Capability) Healthy() bool {
	h := strings.ToUpper(c.ProviderHealth)
	return c.Available() && (h == "HEALTHY" || h == "OK" || h == "SYNCED")
}

// BTCFromRuntime advertises actual BTC microstructure capability. Available != trade.
func BTCFromRuntime(trades, l2, pressure, exhaustion, absorption bool, health string) Capability {
	if health == "" {
		if trades || l2 {
			health = "HEALTHY"
		} else {
			health = "UNKNOWN"
		}
	}
	return Capability{
		Market: "BTC", Trades: trades, IncrementalL2: l2,
		Pressure: pressure, Exhaustion: exhaustion, Absorption: absorption,
		ProviderHealth: health,
	}
}
