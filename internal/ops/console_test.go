package ops

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func waitServer(t *testing.T, s *Server) string {
	t.Helper()
	errCh := make(chan error, 1)
	go func() { errCh <- s.ListenAndServe() }()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if s.Addr() != "" && s.Addr() != "127.0.0.1:0" {
			if resp, err := http.Get("http://" + s.Addr() + "/healthz"); err == nil {
				resp.Body.Close()
				return s.Addr()
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("server")
	return ""
}

func TestConsoleSafetyAndReadOnly(t *testing.T) {
	st := NewStatus()
	st.ExecutionMode = "DEMO"
	st.KillSwitch = true
	st.OpenPositions = 0
	st.LastV1Class = "FLOW_NEUTRAL"
	s := NewServer("127.0.0.1:0", st)
	addr := waitServer(t, s)
	resp, err := http.Get("http://" + addr + "/")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	html := string(body)
	if !strings.Contains(html, "LIVE IMPOSSIBLE") || !strings.Contains(html, "SHADOW") {
		t.Fatal("safety bar")
	}
	for _, bad := range []string{"BUY", "SELL", "CLOSE", "FLATTEN", "CST", "X-SECURITY-TOKEN", "password", "api_key"} {
		if strings.Contains(html, bad) && bad != "CLOSE" {
			t.Fatalf("must not expose %s", bad)
		}
	}
	if strings.Contains(html, "BUY") || strings.Contains(html, "SELL") || strings.Contains(html, "FLATTEN") {
		t.Fatal("mutation controls")
	}
	post, err := http.Post("http://"+addr+"/api/status", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	post.Body.Close()
	if post.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("mutation %d", post.StatusCode)
	}
	gold, err := http.Get("http://" + addr + "/api/gold")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(gold.Body)
	gold.Body.Close()
	if strings.Contains(string(raw), "CST") || strings.Contains(string(raw), "X-SECURITY") {
		t.Fatal("secrets")
	}
}

func TestAPISerializationAndSSE(t *testing.T) {
	s := NewServer("127.0.0.1:0", NewStatus())
	addr := waitServer(t, s)
	for _, p := range []string{"/api/status", "/api/intelligence", "/api/book", "/api/prospective", "/api/research"} {
		resp, err := http.Get("http://" + addr + p)
		if err != nil {
			t.Fatal(p, err)
		}
		var m map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
			t.Fatal(p, err)
		}
		resp.Body.Close()
	}
	req, _ := http.NewRequest(http.MethodGet, "http://"+addr+"/api/stream", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if !strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
		t.Fatal(resp.Header.Get("Content-Type"))
	}
	rd := bufio.NewReader(resp.Body)
	line, err := rd.ReadString('\n')
	if err != nil || !strings.HasPrefix(line, "data:") {
		t.Fatalf("sse %q %v", line, err)
	}
}
