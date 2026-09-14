package mktval

const (
	DataReady             = "DATA_READY"
	ResearchReady         = "RESEARCH_READY"
	ShadowValidated       = "SHADOW_VALIDATED"
	DemoSpecReady         = "DEMO_SPEC_READY"
	DemoCalibrationReady  = "DEMO_CALIBRATION_READY"
	DemoCalibrated        = "DEMO_CALIBRATED"
	DemoEligible          = "DEMO_ELIGIBLE"
	DemoOperational       = "DEMO_OPERATIONAL"
	ResearchRejected      = "RESEARCH_REJECTED"
	Blocked               = "BLOCKED"
)

var Universe = []string{"GOLD", "SILVER", "OIL_CRUDE", "US100", "US500", "US30", "DE40", "UK100", "J225", "CN50"}
