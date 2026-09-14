package mktval

type Row struct {
	Market           string
	Data             string
	History          string
	Mechanics        string
	LegacyCompat     string
	NormCompat       string
	DiscoveryN       int
	ValidationN      int
	HoldoutN         int
	Signals          int
	Expectancy       float64
	Hit              float64
	ProfitFactor     float64
	MaxDD            float64
	CostDrag         float64
	MFE, MAE         float64
	Long, Short      int
	PositiveSubs     int
	AttentionNote    string
	SalienceNote     string
	BrokerSpec       string
	Monetary         string
	Shadow           string
	DemoElig         string
	Trust            string
	Status           string
	DatasetHash      string
	M5, M15, H1, H4  int
	From, To         string
	Spread           float64
	MedianATR        float64
}

func DecideStatus(r Row, hold Stats, dataOK, brokerOK bool) string {
	if !dataOK {
		return ResearchRejected
	}
	if r.M15 < 200 {
		return DataReady
	}
	if hold.N == 0 {
		if r.Signals == 0 {
			return ResearchRejected
		}
		return ResearchReady
	}
	if Promote(hold) && brokerOK {
		return ShadowValidated
	}
	if hold.N >= 8 && hold.Expectancy <= 0 {
		return ResearchRejected
	}
	return ResearchReady
}
