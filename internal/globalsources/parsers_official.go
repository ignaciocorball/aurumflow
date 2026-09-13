package globalsources

import (
	"strings"
	"time"

	"aurumflow/internal/worlddomain"
)

func rewriteFred(id string, raw []byte) []byte {
	var out []byte
	for i, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if i == 0 && (strings.HasPrefix(strings.ToUpper(line), "DATE") || strings.HasPrefix(strings.ToUpper(line), "OBSERVATION")) {
			continue
		}
		cols := splitCSV(line)
		if len(cols) < 2 {
			continue
		}
		out = append(out, []byte(id+","+trim(cols[0])+","+trim(cols[1])+"\n")...)
	}
	return out
}

func ParseFredSeries(seriesID string, raw []byte, retrieved time.Time) []worlddomain.ContextObservation {
	var rebuilt []byte
	rebuilt = append(rebuilt, []byte("SERIES,DATE,VALUE\n")...)
	rebuilt = append(rebuilt, rewriteFred(seriesID, raw)...)
	obs := ParseFedH41(rebuilt, retrieved)
	for i := range obs {
		obs[i].SeriesID = seriesID
		obs[i].DatasetID = "FRED/" + seriesID
		obs[i].SourceURL = "https://fred.stlouisfed.org/series/" + seriesID
	}
	return obs
}

func ParseCboeOfficial(raw []byte, metric string, retrieved time.Time) []worlddomain.ContextObservation {
	var out []worlddomain.ContextObservation
	url := "https://www.cboe.com/us/options/market_statistics/"
	lines := strings.Split(string(raw), "\n")
	if len(lines) == 0 {
		return out
	}
	head := splitCSV(strings.TrimSpace(lines[0]))
	callIdx, putIdx, totIdx, pcIdx, dateIdx := -1, -1, -1, -1, 0
	for i, h := range head {
		u := strings.ToUpper(trim(h))
		switch {
		case u == "DATE" || u == "TRADE DATE":
			dateIdx = i
		case strings.Contains(u, "P/C") || strings.Contains(u, "RATIO") || u == "P_C" || u == "PC":
			pcIdx = i
		case u == "CALL" || u == "CALLS" || u == "CALL VOLUME":
			callIdx = i
		case u == "PUT" || u == "PUTS" || u == "PUT VOLUME":
			putIdx = i
		case u == "TOTAL" || u == "TOTAL VOLUME":
			totIdx = i
		}
	}
	if pcIdx < 0 && len(head) > 1 {
		pcIdx = len(head) - 1
	}
	for _, line := range lines[1:] {
		cols := splitCSV(strings.TrimSpace(line))
		if dateIdx >= len(cols) {
			continue
		}
		obs := parseLooseDate(trim(cols[dateIdx]))
		if obs.IsZero() {
			continue
		}
		avail := obs.AddDate(0, 0, 1)
		add := func(name string, idx int) {
			if idx < 0 || idx >= len(cols) {
				return
			}
			v, ok := parseFloat(cols[idx])
			o := Obs("CBOE", name, v, ok, obs, avail, worlddomain.FreqDaily, worlddomain.RegionUS, worlddomain.AssetEquities, url)
			o.RetrievedAt = retrieved
			o.Period = obs.Format("2006-01-02")
			o.Relation = worlddomain.RelProxy
			o.DatasetID = "cboe_put_call"
			out = append(out, o)
		}
		if metric == "" {
			metric = "TOTAL_PC"
		}
		add(metric, pcIdx)
		if callIdx >= 0 {
			add("CALLS", callIdx)
		}
		if putIdx >= 0 {
			add("PUTS", putIdx)
		}
		if totIdx >= 0 {
			add("TOTAL_VOLUME", totIdx)
		}
	}
	return out
}

func ParseECBSDMX(raw []byte, retrieved time.Time) []worlddomain.ContextObservation {
	var out []worlddomain.ContextObservation
	url := "https://data.ecb.europa.eu/"
	lines := strings.Split(string(raw), "\n")
	if len(lines) == 0 {
		return ParseECB(raw, retrieved)
	}
	head := splitCSV(strings.TrimSpace(lines[0]))
	keyIdx, timeIdx, valIdx := -1, -1, -1
	for i, h := range head {
		u := strings.ToUpper(trim(h))
		switch u {
		case "KEY", "SERIES KEY", "SERIES_KEY":
			keyIdx = i
		case "TIME_PERIOD", "TIME PERIOD", "OBS_DATE":
			timeIdx = i
		case "OBS_VALUE", "OBS VALUE", "VALUE":
			valIdx = i
		}
	}
	if timeIdx < 0 || valIdx < 0 {
		return ParseECB(raw, retrieved)
	}
	for _, line := range lines[1:] {
		cols := splitCSV(strings.TrimSpace(line))
		if timeIdx >= len(cols) || valIdx >= len(cols) {
			continue
		}
		obs := ParseDate(trim(cols[timeIdx]))
		v, ok := parseFloat(cols[valIdx])
		if obs.IsZero() || !ok {
			continue
		}
		key := ""
		if keyIdx >= 0 && keyIdx < len(cols) {
			key = trim(cols[keyIdx])
		}
		metric := "ECB_SERIES"
		switch {
		case strings.Contains(key, "MRR") || strings.Contains(key, "DFR"):
			metric = "POLICY_RATE"
		case strings.Contains(key, "ILM") || strings.Contains(key, "BALANCE"):
			metric = "BALANCE_SHEET"
		}
		o := Obs("ECB", metric, v, true, obs, obs.AddDate(0, 0, 3), worlddomain.FreqWeekly, worlddomain.RegionEurope, worlddomain.AssetFixedInc, url)
		o.RetrievedAt = retrieved
		o.Period = key + "|" + cols[timeIdx]
		o.SeriesID = key
		o.DatasetID = "ECB_SDMX"
		out = append(out, o)
	}
	return out
}

func ParseTICAny(raw []byte, retrieved time.Time) []worlddomain.ContextObservation {
	if strings.Contains(string(raw), "Grand Total") {
		if xs := ParseTICSLTTable1(raw, retrieved); len(xs) > 0 {
			return xs
		}
		if xs := ParseTICSLTTable5(raw, retrieved); len(xs) > 0 {
			return xs
		}
	}
	if xs := ParseTIC(raw, retrieved); len(xs) > 0 {
		return xs
	}
	return ParseTICOfficial(raw, retrieved)
}

func ParseTICOfficial(raw []byte, retrieved time.Time) []worlddomain.ContextObservation {
	var out []worlddomain.ContextObservation
	url := "https://ticdata.treasury.gov/"
	text := string(raw)
	// Official s1_5-style: month columns and a Net row. Extract YYYY-MM tokens plus nearby numbers.
	var months []time.Time
	for _, tok := range strings.Fields(strings.ReplaceAll(text, ",", " ")) {
		if t := parseMonthToken(tok); !t.IsZero() {
			months = append(months, t)
		}
	}
	if len(months) == 0 {
		return out
	}
	var nums []float64
	for _, line := range strings.Split(text, "\n") {
		u := strings.ToUpper(line)
		if !strings.Contains(u, "NET") {
			continue
		}
		for _, col := range strings.Fields(strings.ReplaceAll(line, ",", "")) {
			if v, ok := parseFloat(col); ok {
				nums = append(nums, v)
			}
		}
		if len(nums) > 0 {
			break
		}
	}
	n := len(months)
	if len(nums) < n {
		n = len(nums)
	}
	// Align trailing months to trailing numbers (latest official columns).
	offM := len(months) - n
	offN := len(nums) - n
	for i := 0; i < n; i++ {
		obs := months[offM+i]
		avail := obs.AddDate(0, 2, 0)
		o := Obs("TIC", "NET_LT_SECURITIES", nums[offN+i], true, obs, avail, worlddomain.FreqMonthly, worlddomain.RegionUS, worlddomain.AssetEquities, url)
		o.RetrievedAt = retrieved
		o.Period = obs.Format("2006-01")
		o.DatasetID = "s1_5"
		out = append(out, o)
	}
	return out
}

func ParseICIHTML(raw []byte, retrieved time.Time) []worlddomain.ContextObservation {
	if xs := ParseICI(raw, retrieved); len(xs) > 0 {
		return xs
	}
	text := string(raw)
	// Last-resort: keep UNKNOWN rather than inventing weekly tables from prose.
	if !strings.Contains(strings.ToLower(text), "equity") {
		return nil
	}
	return nil
}

func ParseBISOfficial(raw []byte, retrieved time.Time) []worlddomain.ContextObservation {
	if xs := ParseBIS(raw, retrieved); len(xs) > 0 {
		return xs
	}
	var out []worlddomain.ContextObservation
	url := "https://stats.bis.org/"
	lines := strings.Split(string(raw), "\n")
	if len(lines) == 0 {
		return out
	}
	head := splitCSV(strings.TrimSpace(lines[0]))
	timeIdx, valIdx, ccyIdx := -1, -1, -1
	for i, h := range head {
		u := strings.ToUpper(trim(h))
		switch {
		case u == "TIME_PERIOD" || u == "PERIOD" || u == "TIME":
			timeIdx = i
		case u == "OBS_VALUE" || u == "VALUE" || u == "OBSVALUE":
			valIdx = i
		case u == "CURRENCY" || u == "CCY" || u == "UNIT":
			ccyIdx = i
		}
	}
	if timeIdx < 0 || valIdx < 0 {
		return out
	}
	for _, line := range lines[1:] {
		cols := splitCSV(strings.TrimSpace(line))
		if timeIdx >= len(cols) || valIdx >= len(cols) {
			continue
		}
		obs := ParseDate(trim(cols[timeIdx]))
		v, ok := parseFloat(cols[valIdx])
		if obs.IsZero() || !ok {
			continue
		}
		ccy := "USD"
		if ccyIdx >= 0 && ccyIdx < len(cols) && trim(cols[ccyIdx]) != "" {
			ccy = strings.ToUpper(trim(cols[ccyIdx]))
		}
		o := Obs("BIS", "GLOBAL_CREDIT_"+ccy, v, true, obs, obs.AddDate(0, 2, 0), worlddomain.FreqQuarterly, worlddomain.RegionGlobal, worlddomain.AssetCredit, url)
		o.RetrievedAt = retrieved
		o.Period = trim(cols[timeIdx])
		o.DatasetID = "WS_GLI"
		out = append(out, o)
	}
	return out
}

func ParseJPXAny(raw []byte, retrieved time.Time) []worlddomain.ContextObservation {
	if looksJPXv2(raw) {
		return ParseJPX2026(raw, retrieved)
	}
	if xs := ParseJPX(raw, retrieved); len(xs) > 0 {
		return xs
	}
	return ParseJPX2026(raw, retrieved)
}

func looksJPXv2(raw []byte) bool {
	head := strings.ToUpper(string(raw))
	if len(head) > 400 {
		head = head[:400]
	}
	return strings.Contains(head, "FORMAT") || strings.Contains(head, "POST-APRIL-2026") || strings.Contains(head, "POST_APRIL_2026")
}

func ParseJPX2026(raw []byte, retrieved time.Time) []worlddomain.ContextObservation {
	var out []worlddomain.ContextObservation
	url := "https://www.jpx.co.jp/english/markets/statistics-equities/investor-type/"
	metrics := []string{"FOREIGN", "INDIVIDUALS", "INV_TRUSTS", "CORPORATIONS", "FINANCIALS"}
	for _, line := range strings.Split(string(raw), "\n") {
		cols := splitCSV(strings.TrimSpace(line))
		if len(cols) < 2 {
			continue
		}
		u0 := strings.ToUpper(trim(cols[0]))
		if u0 == "WEEK" || u0 == "DATE" || u0 == "PERIOD" {
			continue
		}
		obs := ParseDate(trim(cols[0]))
		if obs.IsZero() {
			continue
		}
		start := 1
		format := "pre-April-2026"
		if len(cols) > 1 {
			f := strings.ToLower(trim(cols[1]))
			if strings.Contains(f, "post") || strings.Contains(f, "2026") || strings.Contains(f, "v2") {
				format = "post-April-2026"
				start = 2
			}
		}
		avail := obs.AddDate(0, 0, 5)
		for i, name := range metrics {
			idx := start + i
			if idx >= len(cols) {
				break
			}
			v, ok := parseFloat(cols[idx])
			o := Obs("JPX", name, v, ok, obs, avail, worlddomain.FreqWeekly, worlddomain.RegionJapan, worlddomain.AssetEquities, url)
			o.RetrievedAt = retrieved
			o.Unit = "JPY_BN"
			o.DatasetID = "jpx_investor_type"
			o.ParserVer = format
			out = append(out, o)
		}
	}
	return out
}

func ParseTICSLTTable1(raw []byte, retrieved time.Time) []worlddomain.ContextObservation {
	var out []worlddomain.ContextObservation
	url := "https://ticdata.treasury.gov/resource-center/data-chart-center/tic/Documents/slt_table1.txt"
	for _, line := range strings.Split(string(raw), "\n") {
		if !strings.HasPrefix(line, "Grand Total") {
			continue
		}
		cols := strings.Split(line, "\t")
		if len(cols) < 5 {
			cols = strings.Fields(line)
		}
		if len(cols) < 5 {
			continue
		}
		obs := ParseDate(trim(cols[2]))
		hold, hok := parseFloat(cols[3])
		net, nok := parseFloat(cols[4])
		if obs.IsZero() {
			continue
		}
		avail := obs.AddDate(0, 2, 0)
		if hok {
			o := Obs("TIC", "TREASURY_FOREIGN_HOLDINGS", hold/1000, true, obs, avail, worlddomain.FreqMonthly, worlddomain.RegionUS, worlddomain.AssetFixedInc, url)
			o.RetrievedAt = retrieved
			o.Period = obs.Format("2006-01")
			o.Unit = "USD_BN"
			o.DatasetID = "slt_table1"
			out = append(out, o)
		}
		if nok {
			o := Obs("TIC", "NET_LT_SECURITIES", net/1000, true, obs, avail, worlddomain.FreqMonthly, worlddomain.RegionUS, worlddomain.AssetEquities, url)
			o.RetrievedAt = retrieved
			o.Period = obs.Format("2006-01")
			o.Unit = "USD_BN"
			o.DatasetID = "slt_table1"
			out = append(out, o)
		}
	}
	return out
}

func ParseTICSLTTable5(raw []byte, retrieved time.Time) []worlddomain.ContextObservation {
	var out []worlddomain.ContextObservation
	url := "https://ticdata.treasury.gov/resource-center/data-chart-center/tic/Documents/slt_table5.txt"
	var months []time.Time
	for _, line := range strings.Split(string(raw), "\n") {
		cols := strings.Split(strings.TrimSpace(line), "\t")
		if len(cols) < 3 {
			cols = strings.Fields(line)
		}
		if len(cols) > 1 && strings.EqualFold(trim(cols[0]), "Country") {
			for _, c := range cols[1:] {
				if t := ParseDate(trim(c)); !t.IsZero() {
					months = append(months, t)
				}
			}
			continue
		}
		if len(cols) == 0 || !strings.HasPrefix(trim(cols[0]), "Grand Total") || len(months) == 0 {
			continue
		}
		for i, m := range months {
			if i+1 >= len(cols) {
				break
			}
			v, ok := parseFloat(cols[i+1])
			avail := m.AddDate(0, 2, 0)
			o := Obs("TIC", "TREASURY_FOREIGN_HOLDINGS", v, ok, m, avail, worlddomain.FreqMonthly, worlddomain.RegionUS, worlddomain.AssetFixedInc, url)
			o.RetrievedAt = retrieved
			o.Period = m.Format("2006-01")
			o.Unit = "USD_BN"
			o.DatasetID = "slt_table5"
			out = append(out, o)
		}
	}
	return out
}

func ParseFedH41HTML(raw []byte, retrieved time.Time) []worlddomain.ContextObservation {
	text := stripTags(string(raw))
	obs := time.Time{}
	if i := strings.Index(text, "September"); i >= 0 && i+20 < len(text) {
		obs = parseLooseDate(strings.TrimSpace(text[i : i+20]))
	}
	if obs.IsZero() {
		if i := strings.Index(text, "Sep "); i >= 0 && i+15 < len(text) {
			obs = parseLooseDate(strings.TrimSpace(text[i : i+15]))
		}
	}
	if obs.IsZero() {
		obs = retrieved
	}
	var out []worlddomain.ContextObservation
	add := func(label, metric string, min float64) {
		idx := strings.Index(text, label)
		if idx < 0 {
			return
		}
		window := text[idx:]
		if len(window) > 500 {
			window = window[:500]
		}
		for _, tok := range strings.Fields(strings.ReplaceAll(window, ",", "")) {
			v, ok := parseFloat(tok)
			if !ok || v < min {
				continue
			}
			o := Obs("FED_H41", metric, v/1000, true, obs, obs, worlddomain.FreqWeekly, worlddomain.RegionUS, worlddomain.AssetFixedInc, "https://www.federalreserve.gov/releases/h41/current/h41.htm")
			o.RetrievedAt = retrieved
			o.Period = obs.Format("2006-01-02")
			o.Unit = "USD_BN"
			o.DatasetID = "H.4.1"
			o.SeriesID = metric
			out = append(out, o)
			return
		}
	}
	add("Total assets", "FED_ASSETS", 1000000)
	add("Reserve balances with Federal Reserve Banks", "RESERVE_BALANCES", 100000)
	add("U.S. Treasury, General Account", "TGA", 10000)
	add("Reverse repurchase agreements", "ON_RRP", 1000)
	return out
}

func ParseJPXIndexWeeks(raw []byte, retrieved time.Time) []worlddomain.ContextObservation {
	var out []worlddomain.ContextObservation
	url := "https://www.jpx.co.jp/english/markets/statistics-equities/investor-type/index.html"
	seen := map[string]bool{}
	for _, line := range strings.Split(string(raw), "\n") {
		if !strings.Contains(line, "stock_val_1_") {
			continue
		}
		// stock_val_1_260901.xls → 2026-09-01
		idx := strings.Index(line, "stock_val_1_")
		if idx < 0 {
			continue
		}
		code := ""
		for i := idx + len("stock_val_1_"); i < len(line) && i < idx+len("stock_val_1_")+8; i++ {
			if line[i] >= '0' && line[i] <= '9' {
				code += string(line[i])
			} else {
				break
			}
		}
		if len(code) < 6 {
			continue
		}
		obs := ParseDate("20" + code[:6])
		if obs.IsZero() || seen[obs.Format("2006-01-02")] {
			continue
		}
		seen[obs.Format("2006-01-02")] = true
		o := Obs("JPX", "WEEK_PUBLISHED", 1, true, obs, obs.AddDate(0, 0, 5), worlddomain.FreqWeekly, worlddomain.RegionJapan, worlddomain.AssetEquities, url)
		o.RetrievedAt = retrieved
		o.DatasetID = "jpx_investor_type"
		o.ParserVer = "post-April-2026"
		out = append(out, o)
	}
	return out
}

func stripTags(s string) string {
	var b strings.Builder
	in := false
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '<':
			in = true
		case '>':
			in = false
			b.WriteByte(' ')
		default:
			if !in {
				b.WriteByte(s[i])
			}
		}
	}
	return b.String()
}

func parseLooseDate(s string) time.Time {
	if t := ParseDate(s); !t.IsZero() {
		return t
	}
	for _, layout := range []string{"01/02/2006", "1/2/2006", "02-Jan-2006", "2-Jan-2006", "Jan 2, 2006", "January 2, 2006"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}

func parseMonthToken(s string) time.Time {
	s = trim(s)
	if t := ParseDate(s); !t.IsZero() && (len(s) == 7 || strings.Contains(s, "Q")) {
		return t
	}
	if t, err := time.Parse("Jan-06", s); err == nil {
		return t.UTC()
	}
	if t, err := time.Parse("Jan-2006", s); err == nil {
		return t.UTC()
	}
	return time.Time{}
}
