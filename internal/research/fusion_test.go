package research

import "testing"

func TestFusionClassifyAndExplain(t *testing.T) {
	if Classify(1, 1, 40, 15) != AlignedLong {
		t.Fatal("aligned long")
	}
	if Classify(1, -1, -40, 15) != ContradictedLong {
		t.Fatal("contradicted long")
	}
	if Classify(-1, -1, -40, 15) != AlignedShort {
		t.Fatal("aligned short")
	}
	if Classify(1, 0, 5, 15) != LegacyOnly {
		t.Fatal("legacy only")
	}
	if Classify(0, 1, 40, 15) != RadarOnly {
		t.Fatal("radar only")
	}
	s := FusionSnapshot{Class: AlignedLong, LegacyDirection: 1, RadarDirection: 1, RadarPressure: 40,
		LegacyStructure: Layer{Score: 1, Confidence: 0.5, Freshness: "DIRECT", Provenance: "composer"},
		Microstructure:  Layer{Score: 40, Confidence: 0.7, Freshness: "DIRECT", Provenance: "radar"},
	}
	if Explain(s) == "" {
		t.Fatal("explain")
	}
	b := BuildSnapshot(1, 1, 40, 0.8, 15)
	if b.Class != AlignedLong || b.LegacyStructure.Provenance == "" || b.Microstructure.Score != 40 {
		t.Fatalf("%+v", b)
	}
}
