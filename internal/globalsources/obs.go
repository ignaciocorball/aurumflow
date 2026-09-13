package globalsources

import (
	"time"

	"aurumflow/internal/worlddomain"
)

func Obs(source, metric string, value float64, present bool, observed, available time.Time, freq worlddomain.Frequency, region worlddomain.Region, asset worlddomain.AssetClass, url string) worlddomain.ContextObservation {
	return worlddomain.ContextObservation{
		Source: source, Metric: metric, Value: value, Present: present,
		ObservedAt: observed.UTC(), PublishedAt: observed.UTC(), AvailableAt: available.UTC(),
		RetrievedAt: time.Now().UTC(), Frequency: freq, Quality: worlddomain.HealthHealthy,
		Relation: worlddomain.RelSlowContext, Kind: worlddomain.KindObserved,
		Region: region, AssetClass: asset, SourceURL: url,
		Origin: worlddomain.OriginFixture, ParserVer: "v1",
	}
}

func StampOrigin(xs []worlddomain.ContextObservation, origin worlddomain.Origin, hash string) []worlddomain.ContextObservation {
	for i := range xs {
		xs[i].Origin = origin
		if hash != "" {
			xs[i].RawHash = hash
		}
	}
	return xs
}

func ParseDate(s string) time.Time {
	for _, layout := range []string{"2006-01-02", "2006-1-2", "20060102", "2006-01", "2006Q1", time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	if len(s) >= 6 && s[4] == 'Q' {
		y := atoi(s[:4])
		q := atoi(s[5:])
		if y > 0 && q >= 1 && q <= 4 {
			return time.Date(y, time.Month((q-1)*3+1), 1, 0, 0, 0, 0, time.UTC)
		}
	}
	if len(s) >= 7 && s[4] == '-' && s[5] == 'Q' {
		y := atoi(s[:4])
		q := atoi(s[6:])
		if y > 0 && q >= 1 && q <= 4 {
			return time.Date(y, time.Month((q-1)*3+1), 1, 0, 0, 0, 0, time.UTC)
		}
	}
	return time.Time{}
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func parseFloat(s string) (float64, bool) {
	s = trim(s)
	if s == "" || s == "." || s == "NA" || s == "n/a" {
		return 0, false
	}
	neg := false
	if s[0] == '-' {
		neg = true
		s = s[1:]
	}
	dot := -1
	var ip, fp int
	pow := 1.0
	for i, c := range s {
		if c == ',' {
			continue
		}
		if c == '.' {
			if dot >= 0 {
				return 0, false
			}
			dot = i
			continue
		}
		if c < '0' || c > '9' {
			return 0, false
		}
		if dot < 0 {
			ip = ip*10 + int(c-'0')
		} else {
			fp = fp*10 + int(c-'0')
			pow *= 10
		}
	}
	v := float64(ip)
	if pow > 1 {
		v += float64(fp) / pow
	}
	if neg {
		v = -v
	}
	return v, true
}

func trim(s string) string {
	i, j := 0, len(s)
	for i < j && (s[i] == ' ' || s[i] == '\t' || s[i] == '"' || s[i] == '\r') {
		i++
	}
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t' || s[j-1] == '"' || s[j-1] == '\r') {
		j--
	}
	return s[i:j]
}

func splitCSV(line string) []string {
	var out []string
	cur := make([]byte, 0, len(line))
	inQ := false
	for i := 0; i < len(line); i++ {
		c := line[i]
		if c == '"' {
			inQ = !inQ
			continue
		}
		if c == ',' && !inQ {
			out = append(out, string(cur))
			cur = cur[:0]
			continue
		}
		cur = append(cur, c)
	}
	out = append(out, string(cur))
	return out
}

func classifyDelta(d float64, expand, contract float64) worlddomain.LiquidityClass {
	switch {
	case d > expand:
		return worlddomain.LiqExpanding
	case d < contract:
		return worlddomain.LiqContracting
	default:
		return worlddomain.LiqNeutral
	}
}

func pctChange(cur, prev float64) float64 {
	if prev == 0 {
		return 0
	}
	return (cur - prev) / prev * 100
}
