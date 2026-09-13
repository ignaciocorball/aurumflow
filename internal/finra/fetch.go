package finra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const WeeklySummaryURL = "https://api.finra.org/data/group/otcMarket/name/weeklySummary"

type Provider struct {
	HTTP *http.Client
}

func New() *Provider {
	return &Provider{HTTP: &http.Client{Timeout: 20 * time.Second}}
}

// FetchSymbol pulls official weekly OTC summary rows for one equity.
// Missing symbol is not converted to zeros — caller must use Feature() missing flag.
func (p *Provider) FetchSymbol(ctx context.Context, symbol string) ([]Week, error) {
	body := map[string]any{
		"compareFilters": []map[string]string{
			{"compareType": "EQUAL", "fieldName": "issueSymbolIdentifier", "fieldValue": symbol},
		},
		"limit": 52,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, WeeklySummaryURL, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := p.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return nil, fmt.Errorf("FINRA HTTP %d %s", resp.StatusCode, string(b))
	}
	var rows []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}
	out := make([]Week, 0, len(rows))
	for _, r := range rows {
		ending := str(r, "weekStartDate")
		if ending == "" {
			ending = str(r, "summaryDate")
		}
		ats := num(r, "totalWeeklyShareQuantity")
		if v := num(r, "averageWeeklyShareQuantity"); ats == 0 && v != 0 {
			ats = v
		}
		tier := str(r, "tierIdentifier")
		w := ParseWeek(symbol, ending, ats, 0, ending != "")
		if tier != "" {
			w.WeekEnding.Raw = ending + ":" + tier
		}
		out = append(out, w)
	}
	return out, nil
}

func str(m map[string]any, k string) string {
	v, _ := m[k].(string)
	return v
}

func num(m map[string]any, k string) float64 {
	switch v := m[k].(type) {
	case float64:
		return v
	case json.Number:
		f, _ := v.Float64()
		return f
	default:
		return 0
	}
}
