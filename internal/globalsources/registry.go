package globalsources

import (
	"sync"
	"time"

	"aurumflow/internal/worlddomain"
)

type Source struct {
	Name           string
	Coverage       string
	Frequency      worlddomain.Frequency
	ExpectedLag    time.Duration
	OfficialURL    string
	AuthRequired   string
	ParserVersion  string
	LastSuccess    time.Time
	LastAvailable  time.Time
	Health         worlddomain.SensorHealth
	Status         string
}

type Registry struct {
	mu      sync.RWMutex
	sources map[string]Source
}

func NewRegistry() *Registry {
	r := &Registry{sources: map[string]Source{}}
	for _, s := range DefaultSources() {
		r.sources[s.Name] = s
	}
	return r
}

func DefaultSources() []Source {
	return []Source{
		{Name: "BIS", Coverage: "global liquidity, CB assets, policy rates", Frequency: worlddomain.FreqQuarterly, ExpectedLag: 60 * 24 * time.Hour, OfficialURL: "https://www.bis.org/statistics/full_data_sets.htm", AuthRequired: "none", ParserVersion: "v1", Health: worlddomain.HealthUnknown},
		{Name: "FED_H41", Coverage: "US reserve balances, Fed assets, TGA, ON RRP", Frequency: worlddomain.FreqWeekly, ExpectedLag: 5 * 24 * time.Hour, OfficialURL: "https://www.federalreserve.gov/releases/h41/", AuthRequired: "none", ParserVersion: "v1", Health: worlddomain.HealthUnknown},
		{Name: "ECB", Coverage: "policy rates, Eurosystem liquidity", Frequency: worlddomain.FreqWeekly, ExpectedLag: 3 * 24 * time.Hour, OfficialURL: "https://data.ecb.europa.eu/", AuthRequired: "none", ParserVersion: "v1", Health: worlddomain.HealthUnknown},
		{Name: "TIC", Coverage: "US cross-border portfolio and banking", Frequency: worlddomain.FreqMonthly, ExpectedLag: 60 * 24 * time.Hour, OfficialURL: "https://ticdata.treasury.gov/", AuthRequired: "none", ParserVersion: "v1", Health: worlddomain.HealthUnknown},
		{Name: "ICI", Coverage: "US mutual-fund + ETF flow proxy", Frequency: worlddomain.FreqWeekly, ExpectedLag: 4 * 24 * time.Hour, OfficialURL: "https://www.ici.org/research/stats", AuthRequired: "none", ParserVersion: "v1", Health: worlddomain.HealthUnknown},
		{Name: "CFTC", Coverage: "disaggregated COT positioning", Frequency: worlddomain.FreqWeekly, ExpectedLag: 4 * 24 * time.Hour, OfficialURL: "https://www.cftc.gov/dea/newcot/f_disagg.txt", AuthRequired: "none", ParserVersion: "reuse", Health: worlddomain.HealthUnknown},
		{Name: "JPX", Coverage: "Japan investor-type equity flow", Frequency: worlddomain.FreqWeekly, ExpectedLag: 5 * 24 * time.Hour, OfficialURL: "https://www.jpx.co.jp/english/markets/statistics-equities/investor-type/", AuthRequired: "none", ParserVersion: "v1", Health: worlddomain.HealthUnknown},
		{Name: "CBOE", Coverage: "options put/call regime proxy", Frequency: worlddomain.FreqDaily, ExpectedLag: 24 * time.Hour, OfficialURL: "https://www.cboe.com/us/options/market_statistics/", AuthRequired: "none", ParserVersion: "v1", Health: worlddomain.HealthUnknown},
		{Name: "SEC", Coverage: "EDGAR filings / 13F delayed positioning", Frequency: worlddomain.FreqQuarterly, ExpectedLag: 45 * 24 * time.Hour, OfficialURL: "https://data.sec.gov/", AuthRequired: "none", ParserVersion: "reuse", Health: worlddomain.HealthUnknown},
		{Name: "FINRA", Coverage: "OTC weekly ATS context", Frequency: worlddomain.FreqWeekly, ExpectedLag: 7 * 24 * time.Hour, OfficialURL: "https://api.finra.org/", AuthRequired: "none", ParserVersion: "reuse", Health: worlddomain.HealthUnknown},
		{Name: "HKEX", Coverage: "Stock Connect", Frequency: worlddomain.FreqMonthly, OfficialURL: "", AuthRequired: "none", ParserVersion: "v1", Health: worlddomain.HealthUnknown, Status: "PENDING_PUBLIC_STRUCTURED_SOURCE"},
		{Name: "EIA", Coverage: "US petroleum stocks and flows", Frequency: worlddomain.FreqWeekly, OfficialURL: "https://www.eia.gov/", AuthRequired: "free_api_key", ParserVersion: "v1", Health: worlddomain.HealthUnknown, Status: "PENDING_FREE_KEY"},
		{Name: "WGC", Coverage: "gold ETF holdings/flows", Frequency: worlddomain.FreqMonthly, OfficialURL: "https://www.gold.org/", AuthRequired: "none", ParserVersion: "v1", Health: worlddomain.HealthUnknown},
		{Name: "ISHARES", Coverage: "public fund facts watchlist", Frequency: worlddomain.FreqDaily, OfficialURL: "https://www.ishares.com/", AuthRequired: "none", ParserVersion: "v1", Health: worlddomain.HealthUnknown},
		{Name: "EVENT_CONTEXT", Coverage: "future numeric event layer", Frequency: worlddomain.FreqRelease, OfficialURL: "", AuthRequired: "none", ParserVersion: "stub", Health: worlddomain.HealthUnknown, Status: "INTERFACE_ONLY"},
	}
}

func (r *Registry) Get(name string) (Source, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sources[name]
	return s, ok
}

func (r *Registry) All() []Source {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Source, 0, len(r.sources))
	for _, s := range r.sources {
		out = append(out, s)
	}
	return out
}

func (r *Registry) Mark(name string, health worlddomain.SensorHealth, lastAvail, lastOK time.Time, status string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.sources[name]
	s.Name = name
	s.Health = health
	if !lastAvail.IsZero() {
		s.LastAvailable = lastAvail
	}
	if !lastOK.IsZero() {
		s.LastSuccess = lastOK
	}
	if status != "" {
		s.Status = status
	}
	r.sources[name] = s
}
