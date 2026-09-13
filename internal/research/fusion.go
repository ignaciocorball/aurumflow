package research

const (
	LegacyOnly     = "LEGACY_ONLY"
	RadarOnly      = "RADAR_ONLY"
	AlignedLong    = "ALIGNED_LONG"
	AlignedShort   = "ALIGNED_SHORT"
	ContradictedLong  = "CONTRADICTED_LONG"
	ContradictedShort = "CONTRADICTED_SHORT"
	Neutral        = "NEUTRAL"
)

type FusionSnapshot struct {
	Class              string
	LegacyDirection    int
	RadarDirection     int
	RadarPressure      float64
	RadarConfidence    float64
	AlignThreshold     float64
	LegacyStructure    Layer
	Microstructure     Layer
	CrossAsset         Layer
	SlowInstitutional  Layer
	MacroContext       Layer
}

type Layer struct {
	Score      float64
	Confidence float64
	Freshness  string
	Provenance string
}

func BuildSnapshot(legacyDir, radarDir int, pressure, conf, alignAbs float64) FusionSnapshot {
	class := Classify(legacyDir, radarDir, pressure, alignAbs)
	return FusionSnapshot{
		Class: class, LegacyDirection: legacyDir, RadarDirection: radarDir,
		RadarPressure: pressure, RadarConfidence: conf, AlignThreshold: alignAbs,
		LegacyStructure:   Layer{Score: float64(legacyDir), Confidence: 1, Freshness: "DIRECT", Provenance: "composer"},
		Microstructure:    Layer{Score: pressure, Confidence: conf, Freshness: "DIRECT", Provenance: "radar_trade_flow"},
		CrossAsset:        Layer{Freshness: "PROXY_CROSS_ASSET", Provenance: "unavailable"},
		SlowInstitutional: Layer{Freshness: "SLOW_CONTEXT", Provenance: "sec_cftc_finra"},
		MacroContext:      Layer{Freshness: "SLOW_CONTEXT", Provenance: "PENDING_FREE_KEY"},
	}
}

func Classify(legacyDir int, radarDir int, pressure, alignAbs float64) string {
	if alignAbs <= 0 {
		alignAbs = 15
	}
	strong := 0
	if pressure >= alignAbs {
		strong = 1
	} else if pressure <= -alignAbs {
		strong = -1
	}
	if legacyDir == 0 && strong == 0 {
		return Neutral
	}
	if legacyDir == 0 {
		return RadarOnly
	}
	if strong == 0 {
		return LegacyOnly
	}
	if legacyDir > 0 && strong > 0 {
		return AlignedLong
	}
	if legacyDir < 0 && strong < 0 {
		return AlignedShort
	}
	if legacyDir > 0 && strong < 0 {
		return ContradictedLong
	}
	if legacyDir < 0 && strong > 0 {
		return ContradictedShort
	}
	return Neutral
}

func Explain(f FusionSnapshot) string {
	return f.Class +
		" legacy=" + itoa(f.LegacyDirection) +
		" radar=" + itoa(f.RadarDirection) +
		" pressure=" + ftoa(f.RadarPressure)
}

func itoa(n int) string {
	if n > 0 {
		return "+1"
	}
	if n < 0 {
		return "-1"
	}
	return "0"
}

func ftoa(v float64) string {
	return strconvFormat(v)
}
