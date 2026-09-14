package shadowport

import (
	"aurumflow/internal/portfoliorisk"
)

type Trade struct {
	Market string
	Dir    int
	R      float64
	Group  string
}

type Portfolio struct {
	Trades []Trade
}

func (p *Portfolio) Add(t Trade) {
	if t.Group == "" {
		t.Group = portfoliorisk.GroupOf(t.Market)
	}
	p.Trades = append(p.Trades, t)
}

func (p *Portfolio) ReturnR() float64 {
	s := 0.0
	for _, t := range p.Trades {
		s += t.R
	}
	return s
}

func (p *Portfolio) MaxDDR() float64 {
	eq, peak, dd := 0.0, 0.0, 0.0
	for _, t := range p.Trades {
		eq += t.R
		if eq > peak {
			peak = eq
		}
		if peak-eq > dd {
			dd = peak - eq
		}
	}
	return dd
}

func (p *Portfolio) Collisions() int {
	n := 0
	open := map[string]int{}
	for _, t := range p.Trades {
		if open[t.Group] > 0 {
			n++
		}
		open[t.Group]++
	}
	return n
}
