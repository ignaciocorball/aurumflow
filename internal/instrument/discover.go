package instrument

import (
	"strings"

	"aurumflow/internal/worlddomain"
)

type SearchHit struct {
	Epic, Name, Status string
}

func MapHits(hits []SearchHit) []Mapping {
	var out []Mapping
	seen := map[string]bool{}
	for _, h := range hits {
		if strings.TrimSpace(h.Epic) == "" {
			continue
		}
		id, reg, ac := GuessCanonical(h.Name, h.Epic)
		if id == "" {
			continue
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, Mapping{
			Canonical: id, Epic: h.Epic, Name: h.Name,
			Region: string(reg), Asset: string(ac),
		})
	}
	return out
}

func EligibilityFor(id string) worlddomain.ExecEligibility {
	if id == "GOLD" {
		return worlddomain.EligNotCalibrated
	}
	return worlddomain.EligAnalysis
}
