package mktcfg

// Spec holds broker/market mechanics only. Not fitted to returns.
type Spec struct {
	Market         string
	PricePrecision int
	Session        string
	MinM5          int
	MinH1          int
	MinH4          int
	SpreadGateBPS  float64
	ATRNote        string
}

func Of(market string) Spec {
	base := Spec{Market: market, MinM5: 50, MinH1: 20, MinH4: 10, SpreadGateBPS: 8, ATRNote: "do not retune ATR buckets vs returns"}
	switch market {
	case "GOLD":
		base.PricePrecision = 2
		base.Session = "LONDON+NY"
		base.ATRNote = "composer ATR 8-18 is GOLD-scale; compatible != validated"
	case "SILVER":
		base.PricePrecision = 3
		base.Session = "LONDON+NY"
	case "OIL_CRUDE":
		base.PricePrecision = 3
		base.Session = "ALMOST_24H"
	case "US100", "US500", "US30":
		base.PricePrecision = 1
		base.Session = "US_CASH"
	case "DE40", "UK100":
		base.PricePrecision = 1
		base.Session = "EUROPE"
	case "J225", "CN50":
		base.PricePrecision = 1
		base.Session = "ASIA"
	case "BTC":
		base.PricePrecision = 1
		base.Session = "CRYPTO_24X7"
		base.ATRNote = "Legacy unsupported; microstructure lab"
	default:
		base.Session = "UNKNOWN"
	}
	return base
}
