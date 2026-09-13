package crossasset

import (
	"aurumflow/internal/xasset"
	"aurumflow/internal/worlddomain"
)

type Snapshot struct {
	Symbol             string
	Return             float64
	RelStrengthVsUS500 float64
	Vol                float64
	CorrUS500          float64
	Present            bool
}

func FromFrames(frames []xasset.Frame, us500 []float64) []Snapshot {
	var out []Snapshot
	for _, f := range frames {
		if len(f.Close) < 3 {
			out = append(out, Snapshot{Symbol: f.Symbol})
			continue
		}
		rets := xasset.Returns(f.Close)
		s := Snapshot{Symbol: f.Symbol, Present: true, Return: last(rets)}
		s.Vol = vol(rets)
		if len(us500) >= 3 {
			s.RelStrengthVsUS500 = xasset.RelativeStrength(f.Close, us500)
			s.CorrUS500 = xasset.RollingCorr(f.Close, us500, min(20, len(f.Close), len(us500)))
		}
		out = append(out, s)
	}
	return out
}

func MegaCapBreadth(dirs []int) (leaders, laggards int, agreement float64) {
	return xasset.Breadth(dirs)
}

func last(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	return xs[len(xs)-1]
}

func vol(xs []float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	m := 0.0
	for _, x := range xs {
		m += x
	}
	m /= float64(len(xs))
	ss := 0.0
	for _, x := range xs {
		d := x - m
		ss += d * d
	}
	return ss / float64(len(xs)-1)
}

func min(xs ...int) int {
	m := xs[0]
	for _, v := range xs[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

func SignDir(ret float64) int {
	if ret > 0 {
		return 1
	}
	if ret < 0 {
		return -1
	}
	return 0
}

func RegionOf(sym string) worlddomain.Region {
	switch sym {
	case "US100", "US500", "US30":
		return worlddomain.RegionUS
	case "EUROPE", "UK":
		return worlddomain.RegionEurope
	case "JAPAN":
		return worlddomain.RegionJapan
	case "HK", "CHINA":
		return worlddomain.RegionChinaHK
	default:
		return worlddomain.RegionGlobal
	}
}
