package globalsources

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"aurumflow/internal/worlddomain"
)

var (
	RawRoot        = "data/world/raw"
	NormalizedRoot = "data/world/normalized"
)

type CacheMeta struct {
	Provider           string    `json:"provider"`
	OfficialSource     string    `json:"official_source"`
	DatasetID          string    `json:"dataset_id"`
	RetrievedAt        time.Time `json:"retrieved_at"`
	LatestObservation  string    `json:"latest_observation"`
	LatestPublication  string    `json:"latest_publication"`
	Rows               int       `json:"rows"`
	Hash               string    `json:"hash"`
	Origin             string    `json:"origin"`
	Quality            string    `json:"quality"`
	Freshness          string    `json:"freshness"`
	LastError          string    `json:"last_error,omitempty"`
	NextExpected       string    `json:"next_expected_release,omitempty"`
	ParserVersion      string    `json:"parser_version"`
}

func HashBytes(raw []byte) string {
	s := sha256.Sum256(raw)
	return hex.EncodeToString(s[:])
}

func WriteRaw(provider, name string, raw []byte) (string, string, error) {
	dir := filepath.Join(RawRoot, provider)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, raw, 0o644); err != nil {
		return "", "", err
	}
	return p, HashBytes(raw), nil
}

func WriteNormalized(provider string, obs []worlddomain.ContextObservation, meta CacheMeta) error {
	dir := filepath.Join(NormalizedRoot, provider)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(obs, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "observations.json"), body, 0o644); err != nil {
		return err
	}
	mb, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "meta.json"), mb, 0o644)
}

func LoadNormalized(provider string) ([]worlddomain.ContextObservation, CacheMeta, error) {
	var obs []worlddomain.ContextObservation
	var meta CacheMeta
	b, err := os.ReadFile(filepath.Join(NormalizedRoot, provider, "observations.json"))
	if err != nil {
		return nil, meta, err
	}
	if err := json.Unmarshal(b, &obs); err != nil {
		return nil, meta, err
	}
	if mb, err := os.ReadFile(filepath.Join(NormalizedRoot, provider, "meta.json")); err == nil {
		_ = json.Unmarshal(mb, &meta)
	}
	for i := range obs {
		if obs[i].Origin == "" || obs[i].Origin == worlddomain.OriginFixture {
			obs[i].Origin = worlddomain.OriginCache
		}
	}
	return obs, meta, nil
}

func LoadOfficialCache() []worlddomain.ContextObservation {
	var out []worlddomain.ContextObservation
	ents, err := os.ReadDir(NormalizedRoot)
	if err != nil {
		return out
	}
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		xs, _, err := LoadNormalized(e.Name())
		if err != nil {
			continue
		}
		out = append(out, xs...)
	}
	return out
}
