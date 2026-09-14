package mktval

import (
	"math"
	"time"

	"aurumflow/internal/chronosplit"
	"aurumflow/internal/costmodel"
	"aurumflow/internal/indicators"
	"aurumflow/internal/research"
	"aurumflow/pkg/models"
)

type Trade struct {
	T0     time.Time
	Dir    int
	R      float64
	MFE    float64
	MAE    float64
	Hold   int
	Bucket string
}

type Stats struct {
	N, Long, Short int
	Hit            float64
	Expectancy     float64
	MedianR        float64
	PF             float64
	MaxDD          float64
	MFE, MAE       float64
	CostDrag       float64
	PositiveThirds int
	OutlierDom     bool
	LongExp        float64
	ShortExp       float64
	LargestShare   float64
	First, Last    time.Time
}

func Simulate(sigs []research.SignalRow, m15 []models.Candle, spread float64, scenario string, split chronosplit.Split) []Trade {
	var out []Trade
	for _, s := range sigs {
		idx := -1
		for i := range m15 {
			if !m15[i].Time.After(s.Time) {
				idx = i
			}
		}
		if idx < 20 || idx >= len(m15)-1 {
			continue
		}
		atr := indicators.ATR(m15[:idx+1], 14)
		if math.IsNaN(atr) || atr <= 0 {
			continue
		}
		entry := m15[idx].Close
		dir := s.Direction
		if dir == 0 {
			continue
		}
		sl := entry - float64(dir)*atr
		tp := entry + float64(dir)*atr*1.5
		end := idx + 96
		if end > len(m15)-1 {
			end = len(m15) - 1
		}
		exit := m15[end].Close
		hold := end - idx
		mfe, mae := 0.0, 0.0
		for j := idx + 1; j <= end; j++ {
			px := m15[j].Close
			r := float64(dir) * (px - entry) / atr
			if r > mfe {
				mfe = r
			}
			if r < mae {
				mae = r
			}
			hi, lo := m15[j].High, m15[j].Low
			if dir > 0 {
				if lo <= sl {
					exit = sl
					hold = j - idx
					break
				}
				if hi >= tp {
					exit = tp
					hold = j - idx
					break
				}
			} else {
				if hi >= sl {
					exit = sl
					hold = j - idx
					break
				}
				if lo <= tp {
					exit = tp
					hold = j - idx
					break
				}
			}
		}
		net := costmodel.Apply(entry, exit, spread, scenario, dir)
		out = append(out, Trade{
			T0: s.Time, Dir: dir, R: net / atr, MFE: mfe, MAE: mae, Hold: hold,
			Bucket: chronosplit.Bucket(s.Time, split),
		})
	}
	return out
}

func Summarize(tr []Trade) Stats {
	st := Stats{}
	st.N = len(tr)
	if st.N == 0 {
		return st
	}
	var win, lose, sum, mfe, mae, longSum, shortSum float64
	var rs []float64
	eq, peak, dd := 0.0, 0.0, 0.0
	thirds := [3]float64{}
	thirdN := [3]int{}
	best := 0.0
	same := true
	for i := 1; i < len(tr); i++ {
		if tr[i].Bucket != tr[0].Bucket {
			same = false
			break
		}
	}
	for i, t := range tr {
		sum += t.R
		rs = append(rs, t.R)
		if st.First.IsZero() || t.T0.Before(st.First) {
			st.First = t.T0
		}
		if t.T0.After(st.Last) {
			st.Last = t.T0
		}
		if t.Dir > 0 {
			st.Long++
			longSum += t.R
		} else {
			st.Short++
			shortSum += t.R
		}
		if t.R > 0 {
			win += t.R
		} else {
			lose += -t.R
		}
		if t.R > best {
			best = t.R
		}
		mfe += t.MFE
		mae += t.MAE
		eq += t.R
		if eq > peak {
			peak = eq
		}
		if peak-eq > dd {
			dd = peak - eq
		}
		k := 0
		if same {
			k = i * 3 / len(tr)
			if k > 2 {
				k = 2
			}
		} else {
			switch t.Bucket {
			case "VALIDATION":
				k = 1
			case "HOLDOUT":
				k = 2
			}
		}
		thirds[k] += t.R
		thirdN[k]++
	}
	st.Expectancy = sum / float64(st.N)
	st.MFE = mfe / float64(st.N)
	st.MAE = mae / float64(st.N)
	st.MaxDD = dd
	hits := 0
	for _, t := range tr {
		if t.R > 0 {
			hits++
		}
	}
	st.Hit = float64(hits) / float64(st.N)
	if lose > 0 {
		st.PF = win / lose
	} else if win > 0 {
		st.PF = 99
	}
	sortFloat(rs)
	st.MedianR = rs[len(rs)/2]
	for i := 0; i < 3; i++ {
		if thirdN[i] > 0 && thirds[i] >= 0 {
			st.PositiveThirds++
		}
	}
	if st.Long > 0 {
		st.LongExp = longSum / float64(st.Long)
	}
	if st.Short > 0 {
		st.ShortExp = shortSum / float64(st.Short)
	}
	if sum > 0 {
		st.LargestShare = best / sum
	}
	if sum > 0 && best/sum > 0.40 {
		st.OutlierDom = true
	}
	return st
}

func Holdout(tr []Trade) []Trade {
	var out []Trade
	for _, t := range tr {
		if t.Bucket == "HOLDOUT" {
			out = append(out, t)
		}
	}
	return out
}

func Promote(h Stats) bool {
	return h.N >= 20 && h.Expectancy > 0 && h.PF > 1.1 && h.MaxDD < 20 && h.PositiveThirds >= 2 && !h.OutlierDom
}

func sortFloat(xs []float64) {
	for i := 1; i < len(xs); i++ {
		j := i
		for j > 0 && xs[j] < xs[j-1] {
			xs[j], xs[j-1] = xs[j-1], xs[j]
			j--
		}
	}
}
