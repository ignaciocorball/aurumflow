package ops

import "testing"

func TestEvaluatePreflight(t *testing.T) {
	ready := EvaluatePreflight(PreflightInput{
		TestsOK: true, Version: "go1.24", DemoHost: "demo-api-capital.backend-capital.com",
		OpenPositions: 0, GoldStatus: "TRADEABLE", GoldMonetary: "RUNTIME_VALIDATED",
		AccountOK: true, BTCHealth: true, V1Provider: "binance",
		L2Provider: "okx", L2Synced: true, L2Deltas: 10, DiskFreeMB: 1024,
		JournalWritable: true, ProspectiveWritable: true,
	})
	if ready.Result != PreflightReady {
		t.Fatalf("%+v", ready)
	}
	blocked := EvaluatePreflight(PreflightInput{LiveRequested: true, DemoHost: "api-capital.backend-capital.com", OpenPositions: 0})
	if blocked.Result != PreflightBlocked {
		t.Fatalf("%+v", blocked)
	}
	warn := EvaluatePreflight(PreflightInput{
		DemoHost: "demo", OpenPositions: 0, JournalWritable: true, ProspectiveWritable: true, DiskFreeMB: 1024,
	})
	if warn.Result != PreflightReadyWarnings {
		t.Fatalf("%+v", warn)
	}
}
