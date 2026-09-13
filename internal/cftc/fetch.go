package cftc

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const DisaggURL = "https://www.cftc.gov/dea/newcot/f_disagg.txt"

func FetchCurrent(ctx context.Context) ([]Row, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, DisaggURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, err
	}
	return ParseDisagg(resp.Body)
}

func parseF(s string) float64 {
	s = strings.TrimSpace(strings.ReplaceAll(s, ",", ""))
	if s == "" || s == "." {
		return 0
	}
	var v float64
	fmt.Sscanf(s, "%f", &v)
	return v
}

func ParseDisagg(r io.Reader) ([]Row, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	cr.LazyQuotes = true
	var out []Row
	for {
		rec, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return out, err
		}
		if len(rec) < 20 {
			continue
		}
		name := strings.TrimSpace(rec[0])
		if !MatchGold(name) && !strings.Contains(strings.ToUpper(name), "GOLD - COMMODITY EXCHANGE") {
			continue
		}
		asOf, err := time.Parse("2006-01-02", strings.TrimSpace(rec[2]))
		if err != nil {
			asOf, err = time.Parse("01/02/2006", strings.TrimSpace(rec[2]))
			if err != nil {
				continue
			}
		}
		row := Row{Market: GoldContract, AsOf: asOf, Available: AvailableAt(asOf)}
		// Disaggregated futures-only layout (CFTC): after name/date codes,
		// MM long/short typically sit at indices 13/14 when date is YYYY-MM-DD at [2].
		if len(rec) > 17 {
			row.PMNet = parseF(rec[8]) - parseF(rec[9])
			row.SDNet = parseF(rec[10]) - parseF(rec[11])
			row.MMLong = parseF(rec[13])
			row.MMShort = parseF(rec[14])
			row.MMNet = row.MMLong - row.MMShort
			row.ORNet = parseF(rec[16]) - parseF(rec[17])
		}
		out = append(out, row)
	}
	return out, nil
}
