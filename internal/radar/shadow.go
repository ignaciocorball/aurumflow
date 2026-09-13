package radar

// Mode is OFF or SHADOW. AUTO_EXECUTE is intentionally not implemented.
func ValidMode(m string) bool {
	return m == ModeOff || m == ModeShadow || m == ""
}

func MayMutateBroker(mode string) bool {
	return false
}
