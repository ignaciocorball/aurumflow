package globalsources

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

func Catalog() []CacheMeta {
	var out []CacheMeta
	ents, err := os.ReadDir(NormalizedRoot)
	if err != nil {
		return out
	}
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		_, meta, err := LoadNormalized(e.Name())
		if err != nil {
			if mb, rerr := os.ReadFile(filepath.Join(NormalizedRoot, e.Name(), "meta.json")); rerr == nil {
				_ = json.Unmarshal(mb, &meta)
			}
		}
		if meta.Provider == "" {
			meta.Provider = e.Name()
		}
		out = append(out, meta)
	}
	return out
}

func LastActuals() map[string]time.Time {
	out := map[string]time.Time{}
	for _, m := range Catalog() {
		if t := ParseDate(m.LatestPublication); !t.IsZero() {
			out[m.Provider] = t
		} else if t := ParseDate(m.LatestObservation); !t.IsZero() {
			out[m.Provider] = t
		}
	}
	return out
}
