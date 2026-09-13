package quality

type Report struct {
	Duplicates   int `json:"duplicates"`
	OutOfOrder   int `json:"out_of_order"`
	Gaps         int `json:"gaps"`
	InvalidPrice int `json:"invalid_price"`
	InvalidQty   int `json:"invalid_qty"`
	ClockAnom    int `json:"clock_anomalies"`
	Rows         int `json:"rows"`
}

func (r Report) Usable() bool {
	if r.Rows == 0 {
		return false
	}
	if r.InvalidPrice > r.Rows/100 {
		return false
	}
	if r.OutOfOrder > r.Rows/50 {
		return false
	}
	return true
}
