package exhaustion

import "math"

const (
	V1SpecID    = "FLOW_EXHAUSTION_V1"
	V1SpecHash  = "f66e744ec7f8e5785bc6d43b2fc8211941ac725bd463baa33782d955f7399b5d"
	V1Threshold = 15.0

	ClassExhaustion   = "FLOW_EXHAUSTION_CONFIRM"
	ClassContinuation = "FLOW_CONTINUATION_CONFIRM"
	ClassNeutral      = "FLOW_NEUTRAL"
	PotentialExhaustion = "POTENTIAL_FLOW_EXHAUSTION"

	CapTradeFlow uint32 = 1 << iota
	CapBook
)

const (
	ModeShadow     = "SHADOW"
	FeatureVersion = "P5.2-MECH-1"
	AbsorptionNever = "ABSORPTION_CONFIRMED"
)

func MayMutateBroker() bool { return false }

// ClassifyV1 is the frozen V1 rule. Do not retune.
func ClassifyV1(legacyDir int, pressure float64) string {
	return Classify(legacyDir, pressure, V1Threshold)
}

func Classify(legacyDir int, pressure, thresh float64) string {
	if thresh <= 0 {
		thresh = V1Threshold
	}
	if legacyDir == 0 {
		return ClassNeutral
	}
	dp := float64(legacyDir) * pressure
	if dp <= -thresh {
		return ClassExhaustion
	}
	if dp >= thresh {
		return ClassContinuation
	}
	if math.Abs(dp) < thresh {
		return ClassNeutral
	}
	return ClassNeutral
}

func DirectionalPressure(legacyDir int, pressure float64) float64 {
	return float64(legacyDir) * pressure
}
