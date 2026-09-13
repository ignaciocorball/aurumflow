package ops

import "time"

type Status struct {
	UptimeSeconds     int64  `json:"uptime_seconds"`
	APIEnvironment    string `json:"api_environment"`
	ExecutionMode     string `json:"execution_mode"`
	Host              string `json:"host"`
	KillSwitch        bool   `json:"kill_switch"`
	OpenPositions     int    `json:"open_positions"`
	DailyPnL          float64 `json:"daily_pnl"`
	DailyDDPct        float64 `json:"daily_dd_pct"`
	ExecutionEpic     string `json:"execution_epic"`
	MarketStatus      string `json:"market_status"`
	RadarMode         string `json:"radar_mode"`
	RadarProvider     string `json:"radar_provider"`
	BookSynced        bool   `json:"book_synced"`
	EventRate         float64 `json:"event_rate"`
	Reconnects        int    `json:"reconnects"`
	Resyncs           int    `json:"resyncs"`
	CVD               float64 `json:"cvd"`
	Imbalance         float64 `json:"imbalance"`
	Absorption        float64 `json:"absorption"`
	Pressure          float64 `json:"pressure"`
	RadarState        string `json:"radar_state"`
	Confidence        float64 `json:"confidence"`
	LastStrategy      string `json:"last_strategy"`
	LastExecution     string `json:"last_execution"`
	LastError         string `json:"last_error"`
	CME_NQ            string `json:"cme_nq"`
	CME_GC            string `json:"cme_gc"`
	NasdaqTotalView   string `json:"nasdaq_totalview"`
	Options           string `json:"options"`
}

func NewStatus() Status {
	return Status{
		CME_NQ: "NOT_CONNECTED", CME_GC: "NOT_CONNECTED",
		NasdaqTotalView: "NOT_CONNECTED", Options: "NOT_CONNECTED",
		RadarMode: "SHADOW",
	}
}

func AgeSeconds(start time.Time) int64 {
	if start.IsZero() {
		return 0
	}
	return int64(time.Since(start).Seconds())
}
