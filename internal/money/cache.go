package money

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const EvidenceVersion = "v1-capital-cfd-lotsize"

func CachePath(dir, epic string) string {
	epic = strings.ToUpper(strings.TrimSpace(epic))
	if dir == "" {
		dir = filepath.Join("journals", "instruments")
	}
	return filepath.Join(dir, epic+".json")
}

func SaveSpec(path string, s MonetaryInstrumentSpec) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	s.EvidenceVersion = EvidenceVersion
	if s.ValidatedAt.IsZero() {
		s.ValidatedAt = time.Now().UTC()
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func LoadSpec(path string) (MonetaryInstrumentSpec, error) {
	var s MonetaryInstrumentSpec
	data, err := os.ReadFile(path)
	if err != nil {
		return s, err
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return s, err
	}
	return s, nil
}

func MetadataChanged(cached, live MonetaryInstrumentSpec) bool {
	return cached.Epic != live.Epic ||
		cached.Currency != live.Currency ||
		cached.LotSize != live.LotSize ||
		cached.ContractSize != live.ContractSize ||
		cached.ScalingFactor != live.ScalingFactor ||
		cached.MinDealSize != live.MinDealSize ||
		cached.MaxDealSize != live.MaxDealSize ||
		cached.SizeIncrement != live.SizeIncrement ||
		cached.PipPosition != live.PipPosition ||
		cached.TickSize != live.TickSize
}

func InvalidateReason(cached, live MonetaryInstrumentSpec) string {
	if !MetadataChanged(cached, live) {
		return ""
	}
	return fmt.Sprintf("broker metadata changed for %s; revalidation required", live.Epic)
}
