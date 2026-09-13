package research

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type DatasetEntry struct {
	ID     string `json:"id"`
	From   string `json:"from"`
	To     string `json:"to"`
	Status string `json:"status"`
	Role   string `json:"role"`
}

type DatasetRegistry struct {
	Updated  string          `json:"updated_at"`
	Datasets []DatasetEntry  `json:"datasets"`
}

func DefaultDatasetRegistry() DatasetRegistry {
	return DatasetRegistry{
		Updated: time.Now().UTC().Format(time.RFC3339),
		Datasets: []DatasetEntry{
			{ID: "DISCOVERY_DATASET", From: "2026-06-15", To: "2026-09-11", Status: "EXAMINED", Role: "DISCOVERY_ONLY"},
			{ID: "EXTERNAL_HOLDOUT_1", From: "2026-03-17", To: "2026-06-14", Status: "EXAMINED_V1_CONFIRMATORY", Role: "EXTERNAL_HOLDOUT"},
			{ID: "PROSPECTIVE_STREAM", From: "2026-09-13", To: "", Status: "OPEN", Role: "PROSPECTIVE"},
			{ID: "PRE_2026_03_17", From: "", To: "2026-03-16", Status: "UNTOUCHED_VALIDATION_CAPITAL", Role: "RESERVED_FUTURE_VALIDATION"},
		},
	}
}

func LoadDatasetRegistry(path string) (DatasetRegistry, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultDatasetRegistry(), nil
		}
		return DatasetRegistry{}, err
	}
	var r DatasetRegistry
	if err := json.Unmarshal(b, &r); err != nil {
		return DatasetRegistry{}, err
	}
	return r, nil
}

func (r DatasetRegistry) Save(path string) error {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

func (r DatasetRegistry) ForbidUntouched(from, to time.Time) error {
	for _, d := range r.Datasets {
		if d.Status != "UNTOUCHED" && d.Status != "UNTOUCHED_VALIDATION_CAPITAL" {
			continue
		}
		if d.To == "" {
			continue
		}
		end, err := time.Parse("2006-01-02", d.To)
		if err != nil {
			continue
		}
		end = end.Add(24*time.Hour - time.Second)
		if from.Before(end) && (d.From == "" || to.After(mustDay(d.From))) {
			return fmt.Errorf("range overlaps UNTOUCHED dataset %s (%s→%s)", d.ID, d.From, d.To)
		}
	}
	return nil
}

func mustDay(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}
