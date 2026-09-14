package worldstate

import (
	"time"

	"aurumflow/internal/cftc"
	"aurumflow/internal/crossasset"
	"aurumflow/internal/globalsources"
	"aurumflow/internal/livesurface"
	"aurumflow/internal/microcap"
	"aurumflow/internal/sessions"
	"aurumflow/internal/worlddomain"
	"aurumflow/internal/xasset"
)

type Input struct {
	Observations []worlddomain.ContextObservation
	COT          []cftc.Row
	Frames       []xasset.Frame
	US500        []float64
	MegaDirs     []int
	BTCMicro     bool
	GoldQuotes   bool
	Production   bool
	Micro        microcap.Capability
	Live         livesurface.Frame
}

type GlobalLiquidityState struct {
	Class      worlddomain.LiquidityClass
	Label      string
	Fed        worlddomain.LiquidityClass
	ECBDir     string
	BISDir     string
	Confidence float64
	Freshness  worlddomain.Frequency
	Evidence   []worlddomain.SensorEvidence
}

type RatesState struct {
	USPolicy   float64
	USPresent  bool
	EUPolicy   float64
	EUPresent  bool
	Direction  string
}

type CurrencyRegime struct {
	USD        string
	Present    bool
	Evidence   []string
}

type RegionState struct {
	Region     worlddomain.Region
	Equity     string
	Flows      worlddomain.CapitalFlowVector
	Liquidity  string
	Positioning string
	Proxy      string
	Freshness  worlddomain.Frequency
	Health     worlddomain.SensorHealth
}

type AssetClassState struct {
	Class      worlddomain.AssetClass
	Momentum   string
	Flow       string
	Positioning string
	Macro      string
	Risk       string
	Opportunity string
	Health     worlddomain.SensorHealth
}

type MarketState struct {
	Market             string
	Region             worlddomain.Region
	AssetClass         worlddomain.AssetClass
	PriceTrend         string
	RelativeStrength   string
	Volatility         string
	CapitalFlowContext string
	Positioning        string
	MacroAlignment     string
	CrossAsset         string
	MicroAvailable     bool
	DataQuality        worlddomain.SensorHealth
	Eligibility        worlddomain.ExecEligibility
	Setup              worlddomain.SetupState
	Attention          float64
	Coverage           float64
	Tier               string
	Proposal           string
	EligReason         string
	Evidence           []string
	Risks              []string
	Resolved           bool
	MarketStatus       string
	SessionLocal       string
	QuoteAgeSec        float64
	SourceBadge        string
	LegacyCompat       string
	Bid, Ask, Mid      float64
	HistoryStatus      string
	FeatPrice          string
	FeatMomentum       string
	FeatVol            string
	FeatCross          string
	FeatLegacy         string
	TapeState          string
	QuoteHealth        string
	Salience           float64
	SalienceReason     string
}

type WorldState struct {
	AsOf         time.Time
	Session      sessions.Phase
	Liquidity    GlobalLiquidityState
	Risk         worlddomain.RiskRegime
	Rates        RatesState
	USD          CurrencyRegime
	Regions      map[worlddomain.Region]RegionState
	AssetClasses map[worlddomain.AssetClass]AssetClassState
	Markets      map[string]MarketState
	Flows        []worlddomain.CapitalFlowVector
	Opportunity  []MarketState
	Provenance   []worlddomain.SensorEvidence
	Confidence   float64
	Valid        string
	RejectedFix  int
	Origins      []string
	Used         []UsedObs
	Hash         string
	Live         livesurface.Frame
	SlowAt       time.Time
	LiveAt       time.Time
	MicroAt      time.Time
}

func At(t time.Time, in Input) WorldState {
	t = t.UTC()
	obs, rejected := filterProduction(in.Observations, t, in.Production)
	ws := WorldState{
		AsOf: t, Session: sessions.PhaseAt(t, sessions.DefaultSchedule()),
		Regions: map[worlddomain.Region]RegionState{},
		AssetClasses: map[worlddomain.AssetClass]AssetClassState{},
		Markets: map[string]MarketState{},
		Confidence: 1,
	}
	ws.Liquidity = liquidity(obs, t)
	ws.Rates = rates(obs, t)
	ws.USD = usd(obs, t)
	ws.Flows = flows(obs, t)
	xs := crossasset.FromFrames(in.Frames, in.US500)
	leaders, _, agree := crossasset.MegaCapBreadth(in.MegaDirs)
	ws.Risk = risk(obs, t, agree, ws.Liquidity.Class)
	ws.Regions = regions(obs, t, in.COT, xs)
	ws.AssetClasses = assets(obs, t, in.COT, xs, ws.Risk)
	btcMicro := in.BTCMicro || in.Micro.Healthy()
	ws.Markets = markets(obs, t, in.COT, xs, btcMicro, in.GoldQuotes, in.Micro, ws)
	ws.Used = CollectUsed(obs, t)
	if len(in.Live.Features) > 0 || len(in.Live.Quotes) > 0 {
		ws = ApplyLiveFrame(ws, in.Live)
	}
	for _, e := range ws.Liquidity.Evidence {
		ws.Provenance = append(ws.Provenance, e)
	}
	if ws.Liquidity.Fed == worlddomain.LiqUnknown {
		ws.Confidence -= 0.15
	}
	if staleCount(obs, t) > 0 {
		ws.Confidence -= 0.1
	}
	if ws.Confidence < 0.2 {
		ws.Confidence = 0.2
	}
	_ = leaders
	ws.RejectedFix = rejected
	ws.Origins = originList(obs)
	if in.Production && !hasFixture(obs) {
		ws.Valid = "CURRENT_WORLD_STATE_VALID"
	} else if in.Production {
		ws.Valid = "FIXTURE_CONTAMINATED"
	} else {
		ws.Valid = "FIXTURE_OR_TEST"
	}
	ws.SlowAt = t
	return Finalize(ws)
}

func filterProduction(obs []worlddomain.ContextObservation, t time.Time, production bool) ([]worlddomain.ContextObservation, int) {
	rejected := 0
	var out []worlddomain.ContextObservation
	for _, o := range obs {
		if production && !worlddomain.ProductionOrigin(o.Origin) {
			rejected++
			continue
		}
		if o.UsableAt(t) {
			out = append(out, o)
		}
	}
	return out, rejected
}

func hasFixture(obs []worlddomain.ContextObservation) bool {
	for _, o := range obs {
		if o.Origin == worlddomain.OriginFixture {
			return true
		}
	}
	return false
}

func originList(obs []worlddomain.ContextObservation) []string {
	seen := map[string]bool{}
	var out []string
	for _, o := range obs {
		s := string(o.Origin)
		if s == "" {
			s = "UNKNOWN"
		}
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func usable(obs []worlddomain.ContextObservation, t time.Time) []worlddomain.ContextObservation {
	var out []worlddomain.ContextObservation
	for _, o := range obs {
		if o.UsableAt(t) {
			out = append(out, o)
		}
	}
	return out
}

func staleCount(obs []worlddomain.ContextObservation, t time.Time) int {
	n := 0
	for _, o := range obs {
		if o.Quality == worlddomain.HealthStale {
			n++
		}
		if !o.AvailableAt.IsZero() && t.Sub(o.AvailableAt) > 180*24*time.Hour {
			n++
		}
	}
	return n
}

func val(obs []worlddomain.ContextObservation, src, metric string, t time.Time) (worlddomain.ContextObservation, bool) {
	return globalsources.Latest(obs, src, metric, t)
}

func liquidity(obs []worlddomain.ContextObservation, t time.Time) GlobalLiquidityState {
	st := GlobalLiquidityState{Label: "HEURISTIC_V1", Freshness: worlddomain.FreqWeekly, Class: worlddomain.LiqUnknown, Fed: worlddomain.LiqUnknown}
	fed, chg := globalsources.FedLiquidityClass(obs, "FED_ASSETS", t)
	st.Fed = fed
	st.Class = fed
	if _, ok := val(obs, "FED_H41", "FED_ASSETS", t); ok {
		st.Evidence = append(st.Evidence, ev("FED_H41", "FED_ASSETS", "Fed assets observed", t))
		st.Confidence += 0.4
	}
	if o, ok := val(obs, "ECB", "BALANCE_SHEET", t); ok {
		st.ECBDir = dirWord(o.Value)
		st.Evidence = append(st.Evidence, ev("ECB", "BALANCE_SHEET", "Eurosystem balance-sheet observed", t))
		st.Confidence += 0.2
	}
	if o, ok := val(obs, "BIS", "GLOBAL_CREDIT_USD", t); ok {
		st.BISDir = dirWord(o.Value)
		st.Evidence = append(st.Evidence, ev("BIS", "GLOBAL_CREDIT_USD", "USD global credit observed", t))
		st.Confidence += 0.2
	}
	if st.Confidence == 0 {
		st.Class = worlddomain.LiqUnknown
	}
	_ = chg
	return st
}

func rates(obs []worlddomain.ContextObservation, t time.Time) RatesState {
	var r RatesState
	if o, ok := val(obs, "BIS", "POLICY_RATE_USD", t); ok {
		r.USPolicy, r.USPresent = o.Value, true
	}
	if o, ok := val(obs, "ECB", "POLICY_RATE", t); ok {
		r.EUPolicy, r.EUPresent = o.Value, true
	}
	if r.USPresent && r.EUPresent {
		if r.USPolicy > r.EUPolicy {
			r.Direction = "US_PREMIUM"
		} else if r.USPolicy < r.EUPolicy {
			r.Direction = "EU_PREMIUM"
		} else {
			r.Direction = "ALIGNED"
		}
	}
	return r
}

func usd(obs []worlddomain.ContextObservation, t time.Time) CurrencyRegime {
	c := CurrencyRegime{USD: "UNKNOWN"}
	if o, ok := val(obs, "TIC", "NET_LT_SECURITIES", t); ok {
		c.Present = true
		if o.Value > 0 {
			c.USD = "INFLOW_SUPPORT"
		} else if o.Value < 0 {
			c.USD = "OUTFLOW_PRESSURE"
		} else {
			c.USD = "MIXED"
		}
		c.Evidence = append(c.Evidence, "TIC net long-term securities observed")
	}
	return c
}

func flows(obs []worlddomain.ContextObservation, t time.Time) []worlddomain.CapitalFlowVector {
	var out []worlddomain.CapitalFlowVector
	add := func(region worlddomain.Region, asset worlddomain.AssetClass, src, metric string) {
		o, ok := val(obs, src, metric, t)
		v := worlddomain.CapitalFlowVector{Region: region, AssetClass: asset, Direction: worlddomain.FlowUnk, Freshness: worlddomain.FreqWeekly, Health: worlddomain.HealthUnknown}
		if ok && o.Present {
			d, note := globalsources.ICIFlowLabel(metric, o.Value, true)
			if src != "ICI" {
				if o.Value > 0 {
					d = worlddomain.FlowIn
				} else if o.Value < 0 {
					d = worlddomain.FlowOut
				} else {
					d = worlddomain.FlowMixed
				}
				note = src + " " + metric + " observed"
			}
			v.Direction = d
			if o.Value < 0 {
				v.Strength = clamp(-o.Value*10, 0, 100)
			} else {
				v.Strength = clamp(o.Value*10, 0, 100)
			}
			v.Health = worlddomain.HealthHealthy
			v.Evidence = []string{note}
			if src == "ICI" {
				v.Evidence = []string{"fund/ETF capital-flow proxy", metric}
			}
		}
		out = append(out, v)
	}
	add(worlddomain.RegionUS, worlddomain.AssetEquities, "ICI", "DOMESTIC_EQUITY")
	add(worlddomain.RegionGlobal, worlddomain.AssetEquities, "ICI", "WORLD_EQUITY")
	add(worlddomain.RegionUS, worlddomain.AssetFixedInc, "ICI", "BOND")
	add(worlddomain.RegionJapan, worlddomain.AssetEquities, "JPX", "FOREIGN")
	add(worlddomain.RegionGlobal, worlddomain.AssetPrecious, "WGC", "GOLD_ETF_FLOWS")
	return out
}

func risk(obs []worlddomain.ContextObservation, t time.Time, agree float64, liq worlddomain.LiquidityClass) worlddomain.RiskRegime {
	eq, eqok := val(obs, "ICI", "EQUITY", t)
	bond, bok := val(obs, "ICI", "BOND", t)
	pc, pcok := val(obs, "CBOE", "TOTAL_PC", t)
	score := 0
	n := 0
	if eqok && eq.Present {
		n++
		if eq.Value > 0 {
			score++
		} else if eq.Value < 0 {
			score--
		}
	}
	if bok && bond.Present {
		n++
		if bond.Value < 0 {
			score++
		} else if bond.Value > 0 {
			score--
		}
	}
	if pcok && pc.Present {
		n++
		if pc.Value < 1 {
			score++
		} else if pc.Value > 1.2 {
			score--
		}
	}
	if liq == worlddomain.LiqExpanding {
		score++
		n++
	}
	if liq == worlddomain.LiqContracting {
		score--
		n++
	}
	if n == 0 {
		return worlddomain.RiskMixed
	}
	if agree > 0.8 && score > 0 {
		return worlddomain.RiskOn
	}
	if score >= 2 {
		return worlddomain.RiskOn
	}
	if score <= -2 {
		return worlddomain.RiskOff
	}
	if score == 0 {
		return worlddomain.RiskTransition
	}
	return worlddomain.RiskMixed
}

func regions(obs []worlddomain.ContextObservation, t time.Time, cot []cftc.Row, xs []crossasset.Snapshot) map[worlddomain.Region]RegionState {
	out := map[worlddomain.Region]RegionState{}
	for _, r := range []worlddomain.Region{worlddomain.RegionUS, worlddomain.RegionEurope, worlddomain.RegionJapan, worlddomain.RegionChinaHK} {
		st := RegionState{Region: r, Freshness: worlddomain.FreqWeekly, Health: worlddomain.HealthUnknown, Equity: "UNKNOWN", Positioning: "UNKNOWN", Liquidity: "UNKNOWN"}
		switch r {
		case worlddomain.RegionUS:
			if o, ok := val(obs, "ICI", "DOMESTIC_EQUITY", t); ok {
				st.Equity = flowWord(o.Value)
				st.Health = worlddomain.HealthHealthy
			}
			if c := cftc.ContextFor(cot, "NQ", t); c.Present {
				st.Positioning = "CFTC NQ managed-money observed"
			}
		case worlddomain.RegionEurope:
			if _, ok := val(obs, "ECB", "POLICY_RATE", t); ok {
				st.Liquidity = "ECB policy observed"
				st.Health = worlddomain.HealthHealthy
			}
		case worlddomain.RegionJapan:
			if o, ok := val(obs, "JPX", "FOREIGN", t); ok {
				st.Equity = "foreign investor net observed"
				st.Flows = worlddomain.CapitalFlowVector{Region: r, AssetClass: worlddomain.AssetEquities, Direction: signFlow(o.Value), Strength: clamp(abs(o.Value), 0, 100), Health: worlddomain.HealthHealthy, Evidence: []string{"JPX foreign investor net"}}
				st.Health = worlddomain.HealthHealthy
			}
		case worlddomain.RegionChinaHK:
			st.Equity = globalsources.HKEXStatus()
			st.Health = worlddomain.HealthUnknown
		}
		for _, f := range xs {
			if crossasset.RegionOf(f.Symbol) == r && f.Present {
				st.Proxy = f.Symbol
			}
		}
		out[r] = st
	}
	return out
}

func assets(obs []worlddomain.ContextObservation, t time.Time, cot []cftc.Row, xs []crossasset.Snapshot, risk worlddomain.RiskRegime) map[worlddomain.AssetClass]AssetClassState {
	out := map[worlddomain.AssetClass]AssetClassState{}
	for _, a := range []worlddomain.AssetClass{worlddomain.AssetEquities, worlddomain.AssetFixedInc, worlddomain.AssetFX, worlddomain.AssetPrecious, worlddomain.AssetEnergy, worlddomain.AssetCrypto} {
		st := AssetClassState{Class: a, Momentum: "unknown", Flow: "unknown", Positioning: "unknown", Macro: "unknown", Risk: string(risk), Opportunity: "unknown", Health: worlddomain.HealthUnknown}
		switch a {
		case worlddomain.AssetEquities:
			if o, ok := val(obs, "ICI", "EQUITY", t); ok {
				st.Flow = string(signFlow(o.Value))
				st.Health = worlddomain.HealthHealthy
			}
		case worlddomain.AssetFixedInc:
			if o, ok := val(obs, "ICI", "BOND", t); ok {
				st.Flow = string(signFlow(o.Value))
				st.Health = worlddomain.HealthHealthy
			}
		case worlddomain.AssetPrecious:
			if o, ok := val(obs, "WGC", "GOLD_ETF_FLOWS", t); ok {
				st.Flow = string(signFlow(o.Value))
				st.Health = worlddomain.HealthHealthy
			}
			if c := cftc.ContextFor(cot, "GOLD", t); c.Present {
				st.Positioning = "CFTC GOLD managed-money observed"
			}
		case worlddomain.AssetEnergy:
			st.Macro, _ = globalsources.OilPhysicalState(obs, t)
			if st.Macro != "UNKNOWN" {
				st.Health = worlddomain.HealthHealthy
			}
		}
		out[a] = st
	}
	_ = xs
	return out
}

func markets(obs []worlddomain.ContextObservation, t time.Time, cot []cftc.Row, xs []crossasset.Snapshot, btcMicro, goldQuotes bool, micro microcap.Capability, ws WorldState) map[string]MarketState {
	ids := []struct {
		id   string
		reg  worlddomain.Region
		ac   worlddomain.AssetClass
		elig worlddomain.ExecEligibility
	}{
		{"GOLD", worlddomain.RegionGlobal, worlddomain.AssetPrecious, worlddomain.EligAnalysis},
		{"SILVER", worlddomain.RegionGlobal, worlddomain.AssetPrecious, worlddomain.EligAnalysis},
		{"OIL_CRUDE", worlddomain.RegionGlobal, worlddomain.AssetEnergy, worlddomain.EligAnalysis},
		{"US100", worlddomain.RegionUS, worlddomain.AssetEquities, worlddomain.EligAnalysis},
		{"US500", worlddomain.RegionUS, worlddomain.AssetEquities, worlddomain.EligAnalysis},
		{"US30", worlddomain.RegionUS, worlddomain.AssetEquities, worlddomain.EligAnalysis},
		{"DE40", worlddomain.RegionEurope, worlddomain.AssetEquities, worlddomain.EligAnalysis},
		{"UK100", worlddomain.RegionEurope, worlddomain.AssetEquities, worlddomain.EligAnalysis},
		{"J225", worlddomain.RegionJapan, worlddomain.AssetEquities, worlddomain.EligAnalysis},
		{"CN50", worlddomain.RegionChinaHK, worlddomain.AssetEquities, worlddomain.EligAnalysis},
		{"CHINA_HK", worlddomain.RegionChinaHK, worlddomain.AssetEquities, worlddomain.EligAnalysis},
		{"BTC", worlddomain.RegionGlobal, worlddomain.AssetCrypto, worlddomain.EligAnalysis},
	}
	out := map[string]MarketState{}
	xmap := map[string]crossasset.Snapshot{}
	for _, x := range xs {
		xmap[x.Symbol] = x
	}
	for _, m := range ids {
		st := MarketState{Market: m.id, Region: m.reg, AssetClass: m.ac, Eligibility: m.elig, Setup: worlddomain.SetupNone, DataQuality: worlddomain.HealthUnknown, PriceTrend: "UNKNOWN", RelativeStrength: "UNKNOWN", Volatility: "UNKNOWN", CapitalFlowContext: "UNKNOWN", Positioning: "UNKNOWN", MacroAlignment: "UNKNOWN", CrossAsset: "UNKNOWN", Resolved: true, SourceBadge: "UNKNOWN", LegacyCompat: legacyCompat(m.id)}
		if x, ok := xmap[m.id]; ok && x.Present {
			st.PriceTrend = trendWord(x.Return)
			st.RelativeStrength = trendWord(x.RelStrengthVsUS500)
			st.Volatility = "observed"
			st.DataQuality = worlddomain.HealthHealthy
			st.Evidence = append(st.Evidence, "Capital/cross-asset frame observed")
		}
		switch m.id {
		case "GOLD":
			if o, ok := val(obs, "WGC", "GOLD_ETF_FLOWS", t); ok {
				st.CapitalFlowContext = flowWord(o.Value)
				st.Evidence = append(st.Evidence, "Gold ETF flows observed")
				st.DataQuality = worlddomain.HealthHealthy
			}
			if c := cftc.ContextFor(cot, "GOLD", t); c.Present {
				st.Positioning = "CFTC managed-money observed"
				st.Evidence = append(st.Evidence, "CFTC GOLD positioning observed")
			} else if _, ok := val(obs, "CFTC", "GOLD_MM_NET", t); ok {
				st.Positioning = "CFTC managed-money observed"
				st.Evidence = append(st.Evidence, "CFTC GOLD positioning observed")
			}
			st.MicroAvailable = goldQuotes
			st.MacroAlignment = string(ws.Liquidity.Class)
		case "SILVER":
			if c := cftc.ContextFor(cot, "SILVER", t); c.Present {
				st.Positioning = "CFTC silver observed"
			} else if _, ok := val(obs, "CFTC", "SILVER_MM_NET", t); ok {
				st.Positioning = "CFTC silver observed"
			}
		case "OIL_CRUDE", "OIL":
			st.MacroAlignment, _ = globalsources.OilPhysicalState(obs, t)
			if st.MacroAlignment == "" {
				st.MacroAlignment = "UNKNOWN"
			}
			if c := cftc.ContextFor(cot, "CRUDE_OIL", t); c.Present {
				st.Positioning = "CFTC crude observed"
			} else if _, ok := val(obs, "CFTC", "CRUDE_OIL_MM_NET", t); ok {
				st.Positioning = "CFTC crude observed"
			}
		case "US100", "US500", "US30":
			if o, ok := val(obs, "ICI", "DOMESTIC_EQUITY", t); ok {
				st.CapitalFlowContext = flowWord(o.Value)
			}
			if c := cftc.ContextFor(cot, "NQ", t); c.Present && m.id == "US100" {
				st.Positioning = "CFTC NQ observed"
			} else if _, ok := val(obs, "CFTC", "NQ_MM_NET", t); ok && m.id == "US100" {
				st.Positioning = "CFTC NQ observed"
			} else if _, ok := val(obs, "CFTC", "ES_MM_NET", t); ok && m.id == "US500" {
				st.Positioning = "CFTC ES observed"
			}
		case "J225", "JAPAN":
			if o, ok := val(obs, "JPX", "FOREIGN", t); ok {
				st.CapitalFlowContext = flowWord(o.Value)
			}
		case "DE40", "UK100", "EUROPE":
			st.MacroAlignment = string(ws.Rates.Direction)
			if st.MacroAlignment == "" {
				st.MacroAlignment = "UNKNOWN"
			}
		case "CHINA_HK", "CN50":
			st.CapitalFlowContext = globalsources.HKEXStatus()
		case "BTC":
			st.MicroAvailable = btcMicro || (micro.Market == "BTC" && micro.Available())
			if st.MicroAvailable {
				st.Evidence = append(st.Evidence, "BTC microstructure capability: sensing only, not a trade signal")
			}
			st.Evidence = append(st.Evidence, "BTC is MICROSTRUCTURE_LAB + 24/7 SENSOR")
		}
		out[m.id] = st
	}
	return out
}

func ev(src, metric, text string, t time.Time) worlddomain.SensorEvidence {
	return worlddomain.SensorEvidence{Source: src, Metric: metric, Text: text, Kind: worlddomain.KindObserved, Health: worlddomain.HealthHealthy, AsOf: t}
}

func dirWord(v float64) string {
	if v > 0 {
		return "UP"
	}
	if v < 0 {
		return "DOWN"
	}
	return "FLAT"
}

func flowWord(v float64) string {
	if v > 0 {
		return "INFLOW"
	}
	if v < 0 {
		return "OUTFLOW"
	}
	return "MIXED"
}

func signFlow(v float64) worlddomain.FlowDir {
	if v > 0 {
		return worlddomain.FlowIn
	}
	if v < 0 {
		return worlddomain.FlowOut
	}
	return worlddomain.FlowMixed
}

func trendWord(v float64) string {
	if v > 0 {
		return "UP"
	}
	if v < 0 {
		return "DOWN"
	}
	return "FLAT"
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func legacyCompat(id string) string {
	switch id {
	case "GOLD":
		return "LEGACY_COMPATIBLE"
	case "SILVER", "OIL_CRUDE", "US100", "US500", "US30", "DE40", "UK100", "J225", "CN50":
		return "LEGACY_NEEDS_CONFIG"
	default:
		return "LEGACY_NOT_VALIDATED"
	}
}
