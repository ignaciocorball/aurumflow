package optrust

const (
	Trusted     = "DEMO_OPERATIONALLY_TRUSTED"
	Discovered  = "DEMO_DISCOVERED"
	SpecValid   = "DEMO_SPEC_VALID"
	NotTrusted  = "NOT_TRUSTED"
)

type Gate struct {
	Identity     bool
	MarketData   bool
	History      bool
	Strategy     bool
	Monetary     bool
	Risk         bool
	Lifecycle    bool
	Reconcile    bool
}

func Score(g Gate) (string, int, []string) {
	var miss []string
	n := 0
	check := func(ok bool, name string) {
		if ok {
			n++
		} else {
			miss = append(miss, name)
		}
	}
	check(g.Identity, "broker identity")
	check(g.MarketData, "stable market data")
	check(g.History, "history warmup")
	check(g.Strategy, "strategy compatibility")
	check(g.Monetary, "monetary calibration")
	check(g.Risk, "risk computation")
	check(g.Lifecycle, "position lifecycle")
	check(g.Reconcile, "reconciliation")
	if n == 8 {
		return Trusted, 100, nil
	}
	return NotTrusted, n * 12, miss
}

func ForMarket(market string, identity, specValid, calibrated, tradeable bool) (string, Gate) {
	g := Gate{Identity: identity, Strategy: market == "GOLD"}
	if specValid {
		g.MarketData = true
	}
	if calibrated && market == "GOLD" {
		g.Monetary = true
	}
	label, _, _ := Score(g)
	if label == Trusted {
		return Trusted, g
	}
	if specValid {
		return SpecValid, g
	}
	if identity {
		return Discovered, g
	}
	return NotTrusted, g
}
