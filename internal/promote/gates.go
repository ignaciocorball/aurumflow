package promote

import (
	"strings"
	"time"

	"aurumflow/internal/execacct"
	"aurumflow/internal/mktval"
	"aurumflow/internal/money"
)

const (
	FrozenStrategy     = "LEGACY_NORMALIZED_V0"
	MinHoldoutN        = 20
	MinProfitFactor    = 1.1
	MaxDrawdownR       = 20.0
	MinPositiveThirds  = 2
	StressCatastrophic = -0.50
	StressCollapsePF   = 0.75
	StressMaxDDR       = 25.0
)

type Evidence struct {
	Market              string
	DatasetHash         string
	HoldoutStart        time.Time
	HoldoutEnd          time.Time
	HoldoutN            int
	NormalExp           float64
	StressExp           float64
	Hit                 float64
	ProfitFactor        float64
	MaxDD               float64
	MFE                 float64
	MAE                 float64
	LongExp             float64
	ShortExp            float64
	PositiveThirds      int
	LargestShare        float64
	DataQuality         string
	BrokerIdentity      string
	BrokerSpec          string
	Monetary            string
	ResearchStatus      string
	Spread              float64
	Signals             int
	OutlierDominated    bool
	LiveShadow          string
	DemoStatus          string
}

type GateInput struct {
	Evidence       Evidence
	Stress         mktval.Stats
	Normal         mktval.Stats
	Account        execacct.Identity
	WantAccountID  string
	Live           bool
	Env            string
	MonetarySpec   money.MonetaryInstrumentSpec
	BrokerIdentity string
	BrokerUsable   bool
	TradeableCFD   bool
	DataQuality    string
}

type Decision struct {
	Pass     bool
	Status   string
	Failures []string
}

func StressOK(normal, stress mktval.Stats) bool {
	if stress.N == 0 {
		return false
	}
	if stress.Expectancy <= StressCatastrophic {
		return false
	}
	if stress.MaxDD > StressMaxDDR {
		return false
	}
	if normal.Expectancy > 0 && stress.Expectancy < 0 && stress.PF < StressCollapsePF {
		return false
	}
	return true
}

func Evaluate(in GateInput) Decision {
	d := Decision{Status: mktval.ShadowValidated}
	fail := func(s string) { d.Failures = append(d.Failures, s) }
	if in.Live || strings.EqualFold(in.Env, "live") {
		fail("LIVE account rejected")
	}
	if !in.Account.MayTrade || !in.Account.Verified || in.Account.AccountID == "" {
		fail("explicit DEMO account not VERIFIED")
	}
	if in.WantAccountID != "" && in.Account.AccountID != in.WantAccountID {
		fail("DEMO account mismatch")
	}
	if in.Normal.N < MinHoldoutN {
		fail("holdout n insufficient")
	}
	if in.Normal.Expectancy <= 0 {
		fail("net holdout expectancy not > 0 after NORMAL costs")
	}
	if !StressOK(in.Normal, in.Stress) {
		fail("catastrophic reversal under STRESS costs")
	}
	if in.Normal.PF <= MinProfitFactor && in.Normal.PF <= 1 {
		fail("profit factor <= 1")
	} else if in.Normal.PF <= MinProfitFactor {
		fail("profit factor below frozen research contract")
	}
	if in.Normal.MaxDD >= MaxDrawdownR {
		fail("drawdown not acceptable")
	}
	if in.Normal.PositiveThirds < MinPositiveThirds {
		fail("subperiod robustness insufficient")
	}
	if in.Normal.OutlierDom {
		fail("result dominated by one trade")
	}
	dq := strings.ToUpper(strings.TrimSpace(in.DataQuality))
	if dq != "" && dq != "VALID" && dq != "DIRECT" && dq != "OK" {
		fail("data quality not VALID")
	}
	if strings.TrimSpace(in.BrokerIdentity) == "" {
		fail("broker identity unconfirmed")
	}
	if !in.BrokerUsable {
		fail("broker spec not usable")
	}
	if !in.TradeableCFD {
		fail("market not tradeable")
	}
	if in.MonetarySpec.ValidationStatus != money.RuntimeValidated || !money.EvidenceComplete(in.MonetarySpec) {
		fail("RUNTIME_VALIDATED monetary evidence incomplete")
	}
	if len(d.Failures) == 0 {
		d.Pass = true
		d.Status = mktval.DemoEligible
	}
	return d
}

func ResearchPromote(h mktval.Stats) bool {
	return mktval.Promote(h)
}
