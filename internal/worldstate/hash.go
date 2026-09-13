package worldstate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"time"

	"aurumflow/internal/worlddomain"
)

const Version = "WORLDSTATE_V1"
const HashVersion = "WORLD_HASH_V1"

type UsedObs struct {
	Source      string  `json:"source"`
	Metric      string  `json:"metric"`
	Origin      string  `json:"origin"`
	AvailableAt string  `json:"available_at"`
	RetrievedAt string  `json:"retrieved_at"`
	Value       float64 `json:"value"`
}

type hashRegion struct {
	Region      string `json:"region"`
	Equity      string `json:"equity"`
	Flows       string `json:"flows"`
	Liquidity   string `json:"liquidity"`
	Positioning string `json:"positioning"`
	Health      string `json:"health"`
}

type hashAsset struct {
	Class       string `json:"class"`
	Momentum    string `json:"momentum"`
	Flow        string `json:"flow"`
	Positioning string `json:"positioning"`
	Macro       string `json:"macro"`
	Risk        string `json:"risk"`
}

type hashMarket struct {
	Market           string `json:"market"`
	PriceTrend       string `json:"price_trend"`
	RelativeStrength string `json:"relative_strength"`
	Volatility       string `json:"volatility"`
	CapitalFlow      string `json:"capital_flow"`
	Positioning      string `json:"positioning"`
	MacroAlignment   string `json:"macro"`
	CrossAsset       string `json:"cross_asset"`
	MicroAvailable   bool   `json:"micro_available"`
	DataQuality      string `json:"data_quality"`
	Eligibility      string `json:"eligibility"`
}

type hashDoc struct {
	Version    string       `json:"version"`
	AsOf       string       `json:"as_of"`
	Used       []UsedObs    `json:"source_observations"`
	Regions    []hashRegion `json:"regions"`
	Assets     []hashAsset  `json:"asset_classes"`
	Markets    []hashMarket `json:"markets"`
}

func rfc(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
}

func CollectUsed(obs []worlddomain.ContextObservation, t time.Time) []UsedObs {
	var out []UsedObs
	for _, o := range obs {
		if !o.Present || !o.UsableAt(t) {
			continue
		}
		out = append(out, UsedObs{
			Source: o.Source, Metric: o.Metric, Origin: string(o.Origin),
			AvailableAt: rfc(o.AvailableAt), RetrievedAt: rfc(o.RetrievedAt),
			Value: o.Value,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Source != out[j].Source {
			return out[i].Source < out[j].Source
		}
		if out[i].Metric != out[j].Metric {
			return out[i].Metric < out[j].Metric
		}
		return out[i].AvailableAt < out[j].AvailableAt
	})
	return out
}

func Hash(ws WorldState) string {
	doc := hashDoc{Version: HashVersion, AsOf: rfc(ws.AsOf), Used: append([]UsedObs{}, ws.Used...)}
	var rkeys []string
	for k := range ws.Regions {
		rkeys = append(rkeys, string(k))
	}
	sort.Strings(rkeys)
	for _, k := range rkeys {
		r := ws.Regions[worlddomain.Region(k)]
		doc.Regions = append(doc.Regions, hashRegion{
			Region: k, Equity: r.Equity, Flows: string(r.Flows.Direction),
			Liquidity: r.Liquidity, Positioning: r.Positioning, Health: string(r.Health),
		})
	}
	var akeys []string
	for k := range ws.AssetClasses {
		akeys = append(akeys, string(k))
	}
	sort.Strings(akeys)
	for _, k := range akeys {
		a := ws.AssetClasses[worlddomain.AssetClass(k)]
		doc.Assets = append(doc.Assets, hashAsset{
			Class: k, Momentum: a.Momentum, Flow: a.Flow, Positioning: a.Positioning,
			Macro: a.Macro, Risk: a.Risk,
		})
	}
	var mkeys []string
	for k := range ws.Markets {
		mkeys = append(mkeys, k)
	}
	sort.Strings(mkeys)
	for _, k := range mkeys {
		m := ws.Markets[k]
		doc.Markets = append(doc.Markets, hashMarket{
			Market: m.Market, PriceTrend: m.PriceTrend, RelativeStrength: m.RelativeStrength,
			Volatility: m.Volatility, CapitalFlow: m.CapitalFlowContext, Positioning: m.Positioning,
			MacroAlignment: m.MacroAlignment, CrossAsset: m.CrossAsset, MicroAvailable: m.MicroAvailable,
			DataQuality: string(m.DataQuality), Eligibility: string(m.Eligibility),
		})
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func Finalize(ws WorldState) WorldState {
	ws.Hash = Hash(ws)
	return ws
}
