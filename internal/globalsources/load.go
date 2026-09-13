package globalsources

import (
	"os"
	"path/filepath"
	"time"

	"aurumflow/internal/worlddomain"
)

func LoadFixtureDir(dir string, retrieved time.Time) []worlddomain.ContextObservation {
	var out []worlddomain.ContextObservation
	add := func(name string, parse func([]byte, time.Time) []worlddomain.ContextObservation) {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return
		}
		out = append(out, parse(raw, retrieved)...)
	}
	add("bis.csv", ParseBIS)
	add("fed.csv", ParseFedH41)
	add("ecb.csv", ParseECB)
	add("tic.csv", ParseTIC)
	add("ici.csv", ParseICI)
	add("jpx.csv", ParseJPX)
	add("cboe.csv", ParseCboe)
	add("wgc.csv", ParseWGC)
	add("ishares.csv", ParseIShares)
	add("eia.csv", ParseEIA)
	return out
}
