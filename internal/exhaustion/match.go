package exhaustion

import "time"

type MatchKey struct {
	Dir   int
	Hour  int
	VolB  int
	Score int
}

func ScoreBucket(score int) int {
	if score >= 11 {
		return 2
	}
	if score >= 8 {
		return 1
	}
	return 0
}

func VolBucket(rv float64, hist []float64) int {
	if len(hist) < 3 {
		return 1
	}
	cp := append([]float64(nil), hist...)
	sortFloats(cp)
	p33 := quantile(cp, 0.33)
	p66 := quantile(cp, 0.66)
	if rv < p33 {
		return 0
	}
	if rv > p66 {
		return 2
	}
	return 1
}

type MatchRow struct {
	Time   time.Time
	Dir    int
	Hour   int
	Score  int
	Vol    float64
	Class  string
	Idx    int
}

// MatchControls pairs each exhaustion with the nearest unused non-exhaustion
// in the same (dir, hour, vol-bucket, score-bucket). No future outcomes used.
func MatchControls(rows []MatchRow, volHist []float64) (pairs [][2]int) {
	used := map[int]bool{}
	for i, r := range rows {
		if r.Class != ClassExhaustion {
			continue
		}
		want := MatchKey{Dir: r.Dir, Hour: r.Hour, VolB: VolBucket(r.Vol, volHist), Score: ScoreBucket(r.Score)}
		best, bestDt := -1, time.Duration(1<<62)
		for j, c := range rows {
			if used[j] || j == i || c.Class == ClassExhaustion {
				continue
			}
			if c.Dir != want.Dir || c.Hour != want.Hour || ScoreBucket(c.Score) != want.Score {
				continue
			}
			if VolBucket(c.Vol, volHist) != want.VolB {
				continue
			}
			dt := r.Time.Sub(c.Time)
			if dt < 0 {
				dt = -dt
			}
			if dt < bestDt {
				bestDt, best = dt, j
			}
		}
		if best >= 0 {
			used[best] = true
			pairs = append(pairs, [2]int{i, best})
		}
	}
	return pairs
}
