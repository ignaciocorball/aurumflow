package exhaustion

import "time"

type PathPoint struct {
	OffsetMin int     `json:"offset_min"`
	Price     float64 `json:"price_norm"`
	CVD       float64 `json:"cvd_norm"`
	Imbalance float64 `json:"imbalance"`
	Velocity  float64 `json:"trade_velocity"`
}

func BuildPath(minutes []MinuteObs, t0 time.Time) []PathPoint {
	var p0, c0 float64
	found := false
	for _, m := range minutes {
		if !m.T.After(t0) {
			p0, c0 = m.Close, m.CVD
			found = true
		}
	}
	if !found || p0 == 0 {
		return nil
	}
	var out []PathPoint
	for _, m := range minutes {
		off := int(m.T.Sub(t0).Minutes())
		if off < -15 || off > 15 {
			continue
		}
		out = append(out, PathPoint{
			OffsetMin: off,
			Price:     (m.Close - p0) / p0,
			CVD:       m.CVD - c0,
			Imbalance: m.Imbalance,
			Velocity:  m.Velocity,
		})
	}
	return out
}

type MinuteObs struct {
	T         time.Time
	Close     float64
	CVD       float64
	Imbalance float64
	Velocity  float64
}
