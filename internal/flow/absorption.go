package flow

// PotentialAbsorption is a heuristic, not institution identity.
func PotentialAbsorption(aggressiveQty, priceDisplacement, opposingReplenish, typicalQty float64) (bool, string) {
	if typicalQty <= 0 {
		typicalQty = 1
	}
	if aggressiveQty < typicalQty*2 {
		return false, ""
	}
	if priceDisplacement > typicalQty*0.25 {
		return false, ""
	}
	if opposingReplenish < typicalQty {
		return false, ""
	}
	return true, "POTENTIAL_ABSORPTION"
}
