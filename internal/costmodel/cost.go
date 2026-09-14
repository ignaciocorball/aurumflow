package costmodel

const (
	Zero   = "ZERO"
	Normal = "NORMAL"
	Stress = "STRESS"
)

func ExtraSlippage(scenario string, spread float64) float64 {
	if spread < 0 {
		spread = 0
	}
	switch scenario {
	case Stress:
		return 2 * spread
	case Normal:
		return 0.5 * spread
	default:
		return 0
	}
}

func RoundTripCost(spread float64, scenario string) float64 {
	if spread < 0 {
		spread = 0
	}
	return spread + ExtraSlippage(scenario, spread)
}

func Apply(entry, exit, spread float64, scenario string, dir int) float64 {
	cost := RoundTripCost(spread, scenario)
	move := exit - entry
	if dir < 0 {
		move = -move
	}
	return move - cost
}
