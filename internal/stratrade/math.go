package stratrade

import (
	"math"
	"strings"
)

func ExpectedRiskUSD(size, stopDistance, mpu float64) float64 {
	if size <= 0 || stopDistance <= 0 || mpu <= 0 {
		return 0
	}
	return size * stopDistance * mpu
}

func ExpectedUPL(direction string, size, entry, closeable, mpu float64) float64 {
	if size <= 0 || mpu <= 0 {
		return 0
	}
	delta := closeable - entry
	if strings.EqualFold(direction, "SELL") {
		delta = -delta
	}
	return size * delta * mpu
}

func EntrySlippage(direction string, reference, fill float64) float64 {
	if strings.EqualFold(direction, "SELL") {
		return reference - fill
	}
	return fill - reference
}

func ProtectiveOK(expectedSL, expectedTP, brokerSL, brokerTP, tol float64) bool {
	if expectedSL == 0 && expectedTP == 0 {
		return false
	}
	if brokerSL == 0 && brokerTP == 0 {
		return false
	}
	if tol < 0 {
		tol = 0
	}
	if expectedSL != 0 && brokerSL != 0 && math.Abs(expectedSL-brokerSL) > tol {
		return false
	}
	if expectedTP != 0 && brokerTP != 0 && math.Abs(expectedTP-brokerTP) > tol {
		return false
	}
	return true
}

func CountsAgree(local, broker int) bool {
	return local == broker
}

func TradeOutcome(pnl float64, eps float64) string {
	if math.Abs(pnl) <= eps {
		return TradeBreakeven
	}
	if pnl > 0 {
		return TradeWin
	}
	return TradeLoss
}

func OperationalOutcome(mismatch, duplicateOpens, protectiveFail, journalFail bool) string {
	if mismatch || journalFail {
		return OpsFailed
	}
	if duplicateOpens || protectiveFail {
		return OpsWarning
	}
	return OpsClean
}
