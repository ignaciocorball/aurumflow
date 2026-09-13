package globalsources

import (
	"strings"
	"time"

	"aurumflow/internal/worlddomain"
)

func ParseBIS(raw []byte, retrieved time.Time) []worlddomain.ContextObservation {
	var out []worlddomain.ContextObservation
	url := "https://www.bis.org/statistics/full_data_sets.htm"
	for _, line := range strings.Split(string(raw), "\n") {
		cols := splitCSV(strings.TrimSpace(line))
		if len(cols) < 5 || strings.EqualFold(cols[0], "FREQ") {
			continue
		}
		ccy, metric, period := strings.ToUpper(trim(cols[1])), trim(cols[2]), trim(cols[3])
		v, ok := parseFloat(cols[4])
		if !ok {
			continue
		}
		obs := ParseDate(period)
		if obs.IsZero() {
			continue
		}
		avail := obs.AddDate(0, 2, 0)
		region := worlddomain.RegionGlobal
		switch ccy {
		case "USD":
			region = worlddomain.RegionUS
		case "EUR":
			region = worlddomain.RegionEurope
		case "JPY":
			region = worlddomain.RegionJapan
		}
		o := Obs("BIS", strings.ToUpper(metric)+"_"+ccy, v, true, obs, avail, worlddomain.FreqQuarterly, region, worlddomain.AssetCredit, url)
		o.Period = period
		o.RetrievedAt = retrieved
		o.Unit = trim(cols[1])
		out = append(out, o)
	}
	return out
}

func ParseFedH41(raw []byte, retrieved time.Time) []worlddomain.ContextObservation {
	var out []worlddomain.ContextObservation
	url := "https://www.federalreserve.gov/releases/h41/"
	ids := map[string]string{
		"WALCL": "FED_ASSETS", "WRESBAL": "RESERVE_BALANCES",
		"WTREGEN": "TGA", "RRPONTSYD": "ON_RRP",
	}
	for _, line := range strings.Split(string(raw), "\n") {
		cols := splitCSV(strings.TrimSpace(line))
		if len(cols) < 3 || strings.EqualFold(cols[0], "SERIES") {
			continue
		}
		id := strings.ToUpper(trim(cols[0]))
		metric, ok := ids[id]
		if !ok {
			continue
		}
		obs := ParseDate(trim(cols[1]))
		v, pok := parseFloat(cols[2])
		if obs.IsZero() || !pok {
			continue
		}
		o := Obs("FED_H41", metric, v, true, obs, obs.AddDate(0, 0, 5), worlddomain.FreqWeekly, worlddomain.RegionUS, worlddomain.AssetFixedInc, url)
		o.Period = obs.Format("2006-01-02")
		o.RetrievedAt = retrieved
		o.Unit = "USD_BN"
		out = append(out, o)
	}
	return out
}

func FedLiquidityClass(series []worlddomain.ContextObservation, metric string, at time.Time) (worlddomain.LiquidityClass, map[string]float64) {
	xs := Series(series, "FED_H41", metric, at)
	chg := map[string]float64{}
	if len(xs) < 2 {
		return worlddomain.LiqUnknown, chg
	}
	last := xs[len(xs)-1].Value
	look := func(weeks int) float64 {
		cut := xs[len(xs)-1].ObservedAt.AddDate(0, 0, -7*weeks)
		var prev worlddomain.ContextObservation
		ok := false
		for _, o := range xs {
			if !o.ObservedAt.After(cut) {
				prev, ok = o, true
			}
		}
		if !ok {
			return 0
		}
		return pctChange(last, prev.Value)
	}
	chg["1w"] = look(1)
	chg["4w"] = look(4)
	chg["13w"] = look(13)
	chg["52w"] = look(52)
	return classifyDelta(chg["4w"], 0.4, -0.4), chg
}

func ParseECB(raw []byte, retrieved time.Time) []worlddomain.ContextObservation {
	var out []worlddomain.ContextObservation
	url := "https://data.ecb.europa.eu/"
	for _, line := range strings.Split(string(raw), "\n") {
		cols := splitCSV(strings.TrimSpace(line))
		if len(cols) < 3 || strings.EqualFold(cols[0], "KEY") {
			continue
		}
		key := trim(cols[0])
		obs := ParseDate(trim(cols[1]))
		v, ok := parseFloat(cols[2])
		if obs.IsZero() || !ok {
			continue
		}
		metric := "ECB_SERIES"
		switch {
		case strings.Contains(key, "MRR") || strings.Contains(key, "DFR"):
			metric = "POLICY_RATE"
		case strings.Contains(key, "ILM") || strings.Contains(key, "BALANCE"):
			metric = "BALANCE_SHEET"
		case strings.Contains(key, "CREDIT"):
			metric = "CREDIT_CONDITIONS"
		}
		o := Obs("ECB", metric, v, true, obs, obs.AddDate(0, 0, 3), worlddomain.FreqWeekly, worlddomain.RegionEurope, worlddomain.AssetFixedInc, url)
		o.Period = key + "|" + cols[1]
		o.RetrievedAt = retrieved
		out = append(out, o)
	}
	return out
}

func ParseTIC(raw []byte, retrieved time.Time) []worlddomain.ContextObservation {
	var out []worlddomain.ContextObservation
	url := "https://ticdata.treasury.gov/"
	for _, line := range strings.Split(string(raw), "\n") {
		cols := splitCSV(strings.TrimSpace(line))
		if len(cols) < 4 || strings.EqualFold(cols[0], "DATE") {
			continue
		}
		obs := ParseDate(trim(cols[0]))
		if obs.IsZero() {
			continue
		}
		avail := obs.AddDate(0, 2, 0)
		names := []string{"NET_LT_SECURITIES", "FOREIGN_PURCHASES_US", "US_PURCHASES_FOREIGN"}
		for i, name := range names {
			if i+1 >= len(cols) {
				break
			}
			v, ok := parseFloat(cols[i+1])
			o := Obs("TIC", name, v, ok, obs, avail, worlddomain.FreqMonthly, worlddomain.RegionUS, worlddomain.AssetEquities, url)
			o.Period = obs.Format("2006-01")
			o.RetrievedAt = retrieved
			o.Unit = "USD_BN"
			out = append(out, o)
		}
		if len(cols) > 4 {
			if v, ok := parseFloat(cols[4]); ok {
				o := Obs("TIC", "TREASURY_FOREIGN_HOLDINGS", v, true, obs, avail, worlddomain.FreqMonthly, worlddomain.RegionUS, worlddomain.AssetFixedInc, url)
				o.RetrievedAt = retrieved
				out = append(out, o)
			}
		}
		if len(cols) > 5 {
			if v, ok := parseFloat(cols[5]); ok {
				o := Obs("TIC", "CROSS_BORDER_BANKING", v, true, obs, avail, worlddomain.FreqMonthly, worlddomain.RegionUS, worlddomain.AssetCredit, url)
				o.RetrievedAt = retrieved
				out = append(out, o)
			}
		}
	}
	return out
}

func ParseICI(raw []byte, retrieved time.Time) []worlddomain.ContextObservation {
	var out []worlddomain.ContextObservation
	url := "https://www.ici.org/research/stats"
	metrics := []string{"EQUITY", "DOMESTIC_EQUITY", "WORLD_EQUITY", "BOND", "TAXABLE_BOND", "MUNI_BOND", "ETF_ISSUANCE"}
	for _, line := range strings.Split(string(raw), "\n") {
		cols := splitCSV(strings.TrimSpace(line))
		if len(cols) < 2 || strings.Contains(strings.ToLower(cols[0]), "week") {
			continue
		}
		obs := ParseDate(trim(cols[0]))
		if obs.IsZero() {
			continue
		}
		avail := obs.AddDate(0, 0, 4)
		for i, name := range metrics {
			if i+1 >= len(cols) {
				break
			}
			v, ok := parseFloat(cols[i+1])
			asset := worlddomain.AssetEquities
			if strings.Contains(name, "BOND") {
				asset = worlddomain.AssetFixedInc
			}
			o := Obs("ICI", name, v, ok, obs, avail, worlddomain.FreqWeekly, worlddomain.RegionUS, asset, url)
			o.RetrievedAt = retrieved
			o.Unit = "USD_BN"
			o.Kind = worlddomain.KindObserved
			out = append(out, o)
		}
	}
	return out
}

func ICIFlowLabel(metric string, v float64, present bool) (worlddomain.FlowDir, string) {
	if !present {
		return worlddomain.FlowUnk, "UNKNOWN"
	}
	dir := worlddomain.FlowIn
	if v < 0 {
		dir = worlddomain.FlowOut
	} else if v == 0 {
		dir = worlddomain.FlowMixed
	}
	return dir, "fund/ETF capital-flow proxy"
}

func ParseJPX(raw []byte, retrieved time.Time) []worlddomain.ContextObservation {
	var out []worlddomain.ContextObservation
	url := "https://www.jpx.co.jp/english/markets/statistics-equities/investor-type/"
	metrics := []string{"FOREIGN", "INDIVIDUALS", "INV_TRUSTS", "CORPORATIONS", "FINANCIALS"}
	for _, line := range strings.Split(string(raw), "\n") {
		cols := splitCSV(strings.TrimSpace(line))
		if len(cols) < 2 || strings.EqualFold(cols[0], "WEEK") {
			continue
		}
		obs := ParseDate(trim(cols[0]))
		if obs.IsZero() {
			continue
		}
		avail := obs.AddDate(0, 0, 5)
		for i, name := range metrics {
			if i+1 >= len(cols) {
				break
			}
			v, ok := parseFloat(cols[i+1])
			o := Obs("JPX", name, v, ok, obs, avail, worlddomain.FreqWeekly, worlddomain.RegionJapan, worlddomain.AssetEquities, url)
			o.RetrievedAt = retrieved
			o.Unit = "JPY_BN"
			out = append(out, o)
		}
	}
	return out
}

func ParseCboe(raw []byte, retrieved time.Time) []worlddomain.ContextObservation {
	var out []worlddomain.ContextObservation
	url := "https://www.cboe.com/us/options/market_statistics/"
	metrics := []string{"TOTAL_PC", "EQUITY_PC", "INDEX_PC", "SPX_PC", "VIX_PC", "CALLS", "PUTS", "TOTAL_VOLUME"}
	for _, line := range strings.Split(string(raw), "\n") {
		cols := splitCSV(strings.TrimSpace(line))
		if len(cols) < 2 || strings.EqualFold(cols[0], "DATE") {
			continue
		}
		obs := ParseDate(trim(cols[0]))
		if obs.IsZero() {
			continue
		}
		avail := obs.AddDate(0, 0, 1)
		for i, name := range metrics {
			if i+1 >= len(cols) {
				break
			}
			v, ok := parseFloat(cols[i+1])
			o := Obs("CBOE", name, v, ok, obs, avail, worlddomain.FreqDaily, worlddomain.RegionUS, worlddomain.AssetEquities, url)
			o.RetrievedAt = retrieved
			o.Relation = worlddomain.RelProxy
			out = append(out, o)
		}
	}
	return out
}

func ParseWGC(raw []byte, retrieved time.Time) []worlddomain.ContextObservation {
	var out []worlddomain.ContextObservation
	url := "https://www.gold.org/"
	for _, line := range strings.Split(string(raw), "\n") {
		cols := splitCSV(strings.TrimSpace(line))
		if len(cols) < 3 || strings.EqualFold(cols[0], "DATE") {
			continue
		}
		obs := ParseDate(trim(cols[0]))
		hold, hok := parseFloat(cols[1])
		flow, fok := parseFloat(cols[2])
		if obs.IsZero() {
			continue
		}
		avail := obs.AddDate(0, 0, 10)
		o1 := Obs("WGC", "GOLD_ETF_HOLDINGS", hold, hok, obs, avail, worlddomain.FreqMonthly, worlddomain.RegionGlobal, worlddomain.AssetPrecious, url)
		o2 := Obs("WGC", "GOLD_ETF_FLOWS", flow, fok, obs, avail, worlddomain.FreqMonthly, worlddomain.RegionGlobal, worlddomain.AssetPrecious, url)
		o1.RetrievedAt, o2.RetrievedAt = retrieved, retrieved
		out = append(out, o1, o2)
		if len(cols) > 3 {
			if v, ok := parseFloat(cols[3]); ok {
				o := Obs("WGC", "CB_DEMAND", v, true, obs, avail, worlddomain.FreqMonthly, worlddomain.RegionGlobal, worlddomain.AssetPrecious, url)
				o.RetrievedAt = retrieved
				out = append(out, o)
			}
		}
	}
	return out
}

func ParseIShares(raw []byte, retrieved time.Time) []worlddomain.ContextObservation {
	var out []worlddomain.ContextObservation
	url := "https://www.ishares.com/"
	for _, line := range strings.Split(string(raw), "\n") {
		cols := splitCSV(strings.TrimSpace(line))
		if len(cols) < 4 || strings.EqualFold(cols[0], "TICKER") {
			continue
		}
		ticker := strings.ToUpper(trim(cols[0]))
		obs := ParseDate(trim(cols[1]))
		nav, nok := parseFloat(cols[2])
		sh, sok := parseFloat(cols[3])
		if obs.IsZero() {
			continue
		}
		avail := obs.AddDate(0, 0, 1)
		o1 := Obs("ISHARES", ticker+"_AUM", nav, nok, obs, avail, worlddomain.FreqDaily, worlddomain.RegionUS, worlddomain.AssetEquities, url)
		o2 := Obs("ISHARES", ticker+"_SHARES", sh, sok, obs, avail, worlddomain.FreqDaily, worlddomain.RegionUS, worlddomain.AssetEquities, url)
		o1.RetrievedAt, o2.RetrievedAt = retrieved, retrieved
		o1.Kind, o2.Kind = worlddomain.KindObserved, worlddomain.KindObserved
		out = append(out, o1, o2)
	}
	return out
}

func ParseEIA(raw []byte, retrieved time.Time) []worlddomain.ContextObservation {
	var out []worlddomain.ContextObservation
	url := "https://www.eia.gov/"
	metrics := []string{"CRUDE_STOCKS", "SPR", "GASOLINE", "DISTILLATE", "PRODUCTION", "IMPORTS", "EXPORTS", "REFINERY_UTIL"}
	for _, line := range strings.Split(string(raw), "\n") {
		cols := splitCSV(strings.TrimSpace(line))
		if len(cols) < 2 || strings.EqualFold(cols[0], "DATE") {
			continue
		}
		obs := ParseDate(trim(cols[0]))
		if obs.IsZero() {
			continue
		}
		avail := obs.AddDate(0, 0, 2)
		for i, name := range metrics {
			if i+1 >= len(cols) {
				break
			}
			v, ok := parseFloat(cols[i+1])
			o := Obs("EIA", name, v, ok, obs, avail, worlddomain.FreqWeekly, worlddomain.RegionUS, worlddomain.AssetEnergy, url)
			o.RetrievedAt = retrieved
			out = append(out, o)
		}
	}
	return out
}

func OilPhysicalState(obs []worlddomain.ContextObservation, at time.Time) (string, []string) {
	stocks, sok := Latest(obs, "EIA", "CRUDE_STOCKS", at)
	if !sok {
		return "UNKNOWN", []string{"EIA crude stocks unavailable"}
	}
	xs := Series(obs, "EIA", "CRUDE_STOCKS", at)
	why := []string{"commercial crude stocks observed"}
	if len(xs) >= 2 {
		d := xs[len(xs)-1].Value - xs[len(xs)-2].Value
		why = append(why, "weekly delta observed")
		switch {
		case d < 0:
			return "TIGHTENING", why
		case d > 0:
			return "LOOSENING", why
		}
	}
	_ = stocks
	return "BALANCED", why
}

func HKEXStatus() string { return "PENDING_PUBLIC_STRUCTURED_SOURCE" }

type EventContextProvider struct{}

func (EventContextProvider) Name() string { return "EVENT_CONTEXT" }
func (EventContextProvider) Status() string { return "INTERFACE_ONLY" }
