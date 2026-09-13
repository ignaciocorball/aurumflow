package instrument

import "testing"

func TestDefaultRegistry_NoIdenticalCFDFutures(t *testing.T) {
	r := DefaultRegistry()
	goldExec, ok := r.ExecutionFor("GOLD")
	if !ok || goldExec.Symbol != "GOLD" || goldExec.Class != "cfd" {
		t.Fatalf("%+v", goldExec)
	}
	sensors := r.SensorsFor("GOLD")
	if len(sensors) != 1 || sensors[0].Symbol != "GC" || sensors[0].Relation != RelCorrelated {
		t.Fatalf("%+v", sensors)
	}
	btcExec, _ := r.ExecutionFor("BITCOIN")
	btcSens := r.SensorsFor("BITCOIN")[0]
	if SameInstrument(btcExec, btcSens) {
		t.Fatal("Capital BTCUSD CFD is not Binance BTCUSDT")
	}
	if btcSens.Relation == RelIdentical {
		t.Fatal("proxy must not be IDENTICAL")
	}
}
