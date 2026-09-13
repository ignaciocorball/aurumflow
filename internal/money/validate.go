package money

import "math"

func ExpectedPnLDirected(s MonetaryInstrumentSpec, size, openPrice, closePrice float64, direction string) (float64, error) {
	pnl, err := s.ExpectedPnL(size, openPrice, closePrice)
	if err != nil {
		return 0, err
	}
	if direction == "SELL" {
		return -pnl, nil
	}
	return pnl, nil
}

func BrokerUPL(upl, profitLoss float64) float64 {
	if upl != 0 {
		return upl
	}
	return profitLoss
}

func UPLMatches(predicted, broker, absTol, relTol float64) bool {
	if math.IsNaN(predicted) || math.IsNaN(broker) {
		return false
	}
	diff := math.Abs(predicted - broker)
	if diff <= absTol {
		return true
	}
	scale := math.Max(math.Abs(predicted), math.Abs(broker))
	if scale == 0 {
		return true
	}
	return diff/scale <= relTol
}

func RuntimeOK(matches int, samples int) bool {
	return samples >= 2 && matches >= 2
}
