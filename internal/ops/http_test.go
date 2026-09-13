package ops

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestStatusEndpoints(t *testing.T) {
	s := NewServer("127.0.0.1:0", NewStatus())
	s.Set(Status{ExecutionMode: "DEMO", RadarMode: "SHADOW", OpenPositions: 0, LastV1Class: "FLOW_NEUTRAL", BookCapability: "BOOK_CAPABILITY_LIMITED", ImpactFailure: 0.5, FlowEfficiency: -0.2})
	errCh := make(chan error, 1)
	go func() { errCh <- s.ListenAndServe() }()
	deadline := time.Now().Add(2 * time.Second)
	var health *http.Response
	var err error
	for time.Now().Before(deadline) {
		if s.Addr() != "" && s.Addr() != "127.0.0.1:0" {
			health, err = http.Get("http://" + s.Addr() + "/healthz")
			if err == nil {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil || health == nil {
		t.Fatalf("health: %v", err)
	}
	defer health.Body.Close()
	if health.StatusCode != 200 {
		t.Fatal(health.Status)
	}
	resp, err := http.Get("http://" + s.Addr() + "/status")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var st Status
	if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
		t.Fatal(err)
	}
	if st.ExecutionMode != "DEMO" || st.RadarMode != "SHADOW" {
		t.Fatalf("%+v", st)
	}
	if st.LastV1Class != "FLOW_NEUTRAL" || st.BookCapability != "BOOK_CAPABILITY_LIMITED" {
		t.Fatalf("status missing exhaustion fields %+v", st)
	}
}
