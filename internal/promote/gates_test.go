package promote

import (
	"testing"
	"time"

	"aurumflow/internal/chronosplit"
	"aurumflow/internal/costmodel"
	"aurumflow/internal/execacct"
	"aurumflow/internal/mktval"
	"aurumflow/internal/money"
	"aurumflow/internal/research"
	"aurumflow/pkg/models"
)

func demoID() execacct.Identity {
	return execacct.Identity{AccountID: "30demo6430", Verified: true, MayTrade: true, Resolution: execacct.Explicit}
}

func completeSpec() money.MonetaryInstrumentSpec {
	return money.MonetaryInstrumentSpec{
		Epic: "US100", MoneyPerPriceUnit: 1, MoneyPerPriceUnitCurrency: "USD",
		ValidationStatus: money.RuntimeValidated, CalibrationSamples: 80,
		Estimator: money.EstimatorTheilSen, PriceRange: 5.6, UPLRange: 0.0058,
		MetadataStatus: money.MetadataRuntimeAgree, PositionClosed: true, BrokerPositionsAfter: 0,
	}
}

func passStats() mktval.Stats {
	return mktval.Stats{N: 86, Expectancy: 0.152, PF: 1.28, MaxDD: 11.72, PositiveThirds: 3}
}

func passInput() GateInput {
	return GateInput{
		Normal: passStats(),
		Stress: mktval.Stats{N: 86, Expectancy: 0.10, PF: 1.15, MaxDD: 13},
		Account: demoID(), WantAccountID: "30demo6430", Env: "DEMO",
		MonetarySpec: completeSpec(), BrokerIdentity: "capital.com",
		BrokerUsable: true, TradeableCFD: true, DataQuality: "VALID",
	}
}

func TestPromotionGatesPass(t *testing.T) {
	d := Evaluate(passInput())
	if !d.Pass || d.Status != "DEMO_ELIGIBLE" {
		t.Fatalf("%+v", d)
	}
}

func TestPromotionGatesRejectLIVE(t *testing.T) {
	in := passInput()
	in.Live = true
	if Evaluate(in).Pass {
		t.Fatal("LIVE")
	}
	in = passInput()
	in.Env = "live"
	if Evaluate(in).Pass {
		t.Fatal("live env")
	}
}

func TestPromotionGatesRequireVerifiedDemo(t *testing.T) {
	in := passInput()
	in.Account.Verified = false
	if Evaluate(in).Pass {
		t.Fatal("unverified")
	}
	in = passInput()
	in.WantAccountID = "OTHER"
	if Evaluate(in).Pass {
		t.Fatal("mismatch")
	}
}

func TestPromotionGatesRequireRuntimeValidated(t *testing.T) {
	in := passInput()
	in.MonetarySpec.ValidationStatus = money.BrokerMetadataOnly
	in.MonetarySpec.PositionClosed = true
	if Evaluate(in).Pass {
		t.Fatal("metadata only")
	}
	in = passInput()
	in.MonetarySpec.PriceRange = 0
	if Evaluate(in).Pass {
		t.Fatal("incomplete evidence")
	}
}

func TestStressCostGate(t *testing.T) {
	if !StressOK(passStats(), mktval.Stats{N: 86, Expectancy: 0.05, PF: 1.1, MaxDD: 14}) {
		t.Fatal("mild stress should pass")
	}
	if StressOK(passStats(), mktval.Stats{N: 86, Expectancy: -0.8, PF: 0.4, MaxDD: 30}) {
		t.Fatal("catastrophic")
	}
	if StressOK(passStats(), mktval.Stats{N: 86, Expectancy: -0.1, PF: 0.4, MaxDD: 12}) {
		t.Fatal("sign flip with collapsed PF")
	}
	in := passInput()
	in.Stress = mktval.Stats{N: 86, Expectancy: -0.8, PF: 0.3, MaxDD: 30}
	if Evaluate(in).Pass {
		t.Fatal("stress gate")
	}
}

func TestHoldoutEvidenceExtraction(t *testing.T) {
	t0 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	var m15 []models.Candle
	for i := 0; i < 80; i++ {
		px := 100.0 + float64(i)*0.2
		m15 = append(m15, models.Candle{Time: t0.Add(time.Duration(i) * 15 * time.Minute), Open: px, High: px + 1, Low: px - 1, Close: px})
	}
	split := chronosplit.Of(m15[0].Time, m15[len(m15)-1].Time)
	sigs := []research.SignalRow{{Time: m15[70].Time, Direction: 1}}
	all := mktval.Simulate(sigs, m15, 0.2, costmodel.Normal, split)
	if len(all) != 1 {
		t.Fatalf("sim=%+v", all)
	}
	st := mktval.Summarize(mktval.Holdout(all))
	if st.N != 1 && mktval.Summarize(all).N != 1 {
		t.Fatalf("hold=%+v all=%+v bucket=%s", st, all, all[0].Bucket)
	}
	st = mktval.Summarize(all)
	if st.N != 1 || st.First.IsZero() || st.Last.IsZero() {
		t.Fatalf("%+v", st)
	}
	hash := HashCandles(m15)
	if len(hash) != 64 {
		t.Fatal(hash)
	}
}

func TestResearchPromoteDoesNotUseAttention(t *testing.T) {
	if ResearchPromote(mktval.Stats{N: 3, Expectancy: 2, PF: 9, MaxDD: 1, PositiveThirds: 3}) {
		t.Fatal("tiny n")
	}
}
