package money

import (
	"math"
	"testing"
	"time"
)

func synth(dir string, size, entry, mpu, intercept float64, prices []float64, round int, outlierIdx int, outlierUPL float64) []CalibrationSample {
	sign := DirectionSign(dir)
	out := make([]CalibrationSample, 0, len(prices))
	for i, px := range prices {
		closeable := CloseablePrice(dir, px, px+0.5)
		upl := size*sign*(closeable-entry)*mpu + intercept
		if round >= 0 {
			pow := math.Pow(10, float64(round))
			upl = math.Round(upl*pow) / pow
		}
		if i == outlierIdx {
			upl = outlierUPL
		}
		out = append(out, CalibrationSample{
			Timestamp: time.Unix(int64(i), 0).UTC(), DealID: "D1", Direction: dir, Size: size,
			Bid: px, Ask: px + 0.5, CloseablePrice: closeable,
			BrokerUPL: upl, PositionOpenLevel: entry, AccountCurrency: "USD",
		})
	}
	return out
}

func goldCfg(meta float64) CalibrationConfig {
	cfg := DefaultCalibrationConfig()
	cfg.TickSize = 0.01
	cfg.Spread = 0.5
	cfg.MetadataMPU = meta
	return cfg
}

func TestTheilSenBUYKnownSlope(t *testing.T) {
	prices := []float64{4330, 4330.2, 4330.5, 4330.9, 4331.4, 4331.8, 4332.2, 4332.7, 4333.1, 4333.6}
	res := EstimateTheilSen(synth("BUY", 0.01, 4330, 1, 0, prices, -1, -1, 0), goldCfg(1))
	if !res.OK || math.Abs(res.MoneyPerPriceUnit-1) > 0.05 {
		t.Fatalf("%+v", res)
	}
}

func TestTheilSenSELLInverseRawPositiveMPU(t *testing.T) {
	prices := []float64{4330, 4330.3, 4330.7, 4331.1, 4331.6, 4332.0, 4332.5, 4333.0, 4333.4, 4333.9}
	res := EstimateTheilSen(synth("SELL", 0.01, 4330, 1, 0, prices, -1, -1, 0), goldCfg(1))
	if !res.OK || math.Abs(res.MoneyPerPriceUnit-1) > 0.05 {
		t.Fatalf("%+v", res)
	}
}

func TestTheilSenConstantInterceptCancels(t *testing.T) {
	prices := []float64{100, 100.4, 100.9, 101.3, 101.8, 102.2, 102.7, 103.1, 103.6, 104.0}
	a := EstimateTheilSen(synth("BUY", 0.01, 100, 2, 0, prices, -1, -1, 0), goldCfg(0))
	b := EstimateTheilSen(synth("BUY", 0.01, 100, 2, 7.5, prices, -1, -1, 0), goldCfg(0))
	if !a.OK || !b.OK || math.Abs(a.MoneyPerPriceUnit-b.MoneyPerPriceUnit) > 0.02 {
		t.Fatalf("a=%+v b=%+v", a, b)
	}
}

func TestTheilSenRoundedUPL(t *testing.T) {
	prices := []float64{2000, 2000.4, 2000.9, 2001.5, 2002.1, 2002.8, 2003.4, 2004.0, 2004.7, 2005.3}
	res := EstimateTheilSen(synth("BUY", 0.01, 2000, 1, 0, prices, 2, -1, 0), goldCfg(1))
	if !res.OK || math.Abs(res.MoneyPerPriceUnit-1) > 0.15 {
		t.Fatalf("%+v", res)
	}
}

func TestTheilSenResistsOneOutlier(t *testing.T) {
	prices := []float64{100, 100.3, 100.7, 101.1, 101.6, 102.0, 102.4, 102.9, 103.3, 103.8}
	res := EstimateTheilSen(synth("BUY", 0.01, 100, 1, 0, prices, -1, 3, 50), goldCfg(1))
	if !res.OK || math.Abs(res.MoneyPerPriceUnit-1) > 0.2 {
		t.Fatalf("%+v", res)
	}
}

func TestInsufficientPriceExcursion(t *testing.T) {
	prices := []float64{4337.10, 4337.10, 4337.10, 4337.10, 4337.10, 4337.10, 4337.10, 4337.10, 4337.11, 4337.11}
	res := EstimateTheilSen(synth("BUY", 0.01, 4337.10, 1, 0, prices, -1, -1, 0), goldCfg(0))
	if res.OK || res.Failure != FailInsufficientPriceExcursion {
		t.Fatalf("%+v", res)
	}
}

func TestInsufficientUPLResolution(t *testing.T) {
	prices := []float64{10, 10.3, 10.7, 11.1, 11.6, 12.0, 12.5, 13.0, 13.4, 13.9}
	xs := synth("BUY", 0.01, 10, 1, 0, prices, -1, -1, 0)
	for i := range xs {
		xs[i].BrokerUPL = 0
	}
	res := EstimateTheilSen(xs, goldCfg(0))
	if res.OK || res.Failure != FailInsufficientUPLResolution {
		t.Fatalf("%+v", res)
	}
}

func TestWrongSlopeSign(t *testing.T) {
	prices := []float64{100, 100.4, 100.9, 101.4, 101.9, 102.4, 102.9, 103.4, 103.9, 104.4}
	xs := synth("BUY", 0.01, 100, 1, 0, prices, -1, -1, 0)
	for i := range xs {
		xs[i].BrokerUPL = -xs[i].BrokerUPL
	}
	res := EstimateTheilSen(xs, goldCfg(0))
	if res.OK || res.Failure != FailWrongSlopeSign {
		t.Fatalf("%+v", res)
	}
}

func TestMetadataConflict(t *testing.T) {
	prices := []float64{100, 100.4, 100.9, 101.4, 101.9, 102.4, 102.9, 103.4, 103.9, 104.4}
	res := EstimateTheilSen(synth("BUY", 0.01, 100, 1, 0, prices, -1, -1, 0), goldCfg(10))
	if res.OK || res.Failure != FailMonetaryConflict {
		t.Fatalf("%+v", res)
	}
}

func TestApplyRuntimeAndCloseGuard(t *testing.T) {
	spec := MonetaryInstrumentSpec{Epic: "GOLD", ValidationStatus: BrokerMetadataOnly, MoneyPerPriceUnit: 1}
	fail := ApplyRuntime(spec, CalibrationResult{Failure: FailInsufficientSamples})
	if fail.ValidationStatus != BrokerMetadataOnly {
		t.Fatal(fail.ValidationStatus)
	}
	ok := ApplyRuntime(spec, CalibrationResult{OK: true, MoneyPerPriceUnit: 1.02, Currency: "USD", EvidenceDealID: "D1", Samples: 12})
	if ok.ValidationStatus != RuntimeValidated || ok.MoneyPerPriceUnitCurrency != "USD" || ok.CalibrationDealID != "D1" {
		t.Fatalf("%+v", ok)
	}
	n := 0
	g := CloseGuard{Close: func() error { n++; return nil }}
	func() { defer g.Ensure() }()
	if n != 1 || g.Ensure() != nil || n != 1 {
		t.Fatalf("close n=%d", n)
	}
}

func TestCloseablePriceUsesBidAsk(t *testing.T) {
	if CloseablePrice("BUY", 10, 10.5) != 10 {
		t.Fatal("buy")
	}
	if CloseablePrice("SELL", 10, 10.5) != 10.5 {
		t.Fatal("sell")
	}
}
