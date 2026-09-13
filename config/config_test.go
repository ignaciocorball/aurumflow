package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func validDemoAPI() APIConfig {
	return APIConfig{
		Environment:   "demo",
		Mode:          "demo",
		ExecutionMode: "DISABLED",
		APIKey:        "test-key",
		Identifier:    "user@example.com",
		Password:      "test-pass",
	}
}

func TestValidate_MissingEnvironment(t *testing.T) {
	c := &Config{API: validDemoAPI()}
	c.API.Environment = ""
	c.API.Mode = ""
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("missing environment: want refuse, got %v", err)
	}
}

func TestValidate_EmptyEnvironmentWhitespace(t *testing.T) {
	c := &Config{API: validDemoAPI()}
	c.API.Environment = "   "
	c.API.Mode = ""
	if err := c.Validate(); err == nil {
		t.Fatal("empty environment must refuse")
	}
}

func TestValidate_InvalidEnvironment(t *testing.T) {
	c := &Config{API: validDemoAPI()}
	c.API.Environment = "staging"
	c.API.Mode = ""
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("invalid environment: want refuse, got %v", err)
	}
}

func TestValidate_LiveEnvironment(t *testing.T) {
	c := &Config{API: validDemoAPI()}
	c.API.Environment = "live"
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "LIVE") {
		t.Fatalf("LIVE environment: want %q, got %v", ErrLiveDisabled, err)
	}
}

func TestValidate_LiveModeLegacy(t *testing.T) {
	c := &Config{API: validDemoAPI()}
	c.API.Environment = "demo"
	c.API.Mode = "live"
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "LIVE") {
		t.Fatalf("legacy mode=live must refuse, got %v", err)
	}
}

func TestValidate_LiveURLManual(t *testing.T) {
	c := &Config{API: validDemoAPI()}
	c.API.BaseURL = LiveAPIHost
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "LIVE") {
		t.Fatalf("manual LIVE URL must refuse, got %v", err)
	}
}

func TestValidate_DemoDisabledAllowsStartup(t *testing.T) {
	c := &Config{API: validDemoAPI()}
	c.API.ExecutionMode = "DISABLED"
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if c.API.BaseURL != DemoAPIHost {
		t.Fatalf("host=%s", c.API.BaseURL)
	}
	if c.ExecMode() != ExecutionDisabled {
		t.Fatalf("mode=%s", c.ExecMode())
	}
}

func TestValidate_DemoDryRun(t *testing.T) {
	c := &Config{API: validDemoAPI()}
	c.API.ExecutionMode = "DRY_RUN"
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if c.ExecMode() != ExecutionDryRun {
		t.Fatalf("mode=%s", c.ExecMode())
	}
}

func TestValidate_DemoExecution(t *testing.T) {
	c := &Config{API: validDemoAPI()}
	c.API.ExecutionMode = "DEMO"
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if c.ExecMode() != ExecutionDemo {
		t.Fatalf("mode=%s", c.ExecMode())
	}
}

func TestValidate_LiveExecutionMode(t *testing.T) {
	c := &Config{API: validDemoAPI()}
	c.API.ExecutionMode = "LIVE"
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "LIVE") {
		t.Fatalf("execution_mode=LIVE must refuse, got %v", err)
	}
}

func TestValidate_ArbitraryHostRefused(t *testing.T) {
	c := &Config{API: validDemoAPI()}
	c.API.BaseURL = "https://evil.example.com"
	if err := c.Validate(); err == nil {
		t.Fatal("arbitrary host must refuse")
	}
}

func TestValidate_DemoHostForced(t *testing.T) {
	c := &Config{API: validDemoAPI()}
	c.API.BaseURL = DemoAPIHost + "/extra"
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if c.API.BaseURL != DemoAPIHost {
		t.Fatalf("expected pinned demo host, got %s", c.API.BaseURL)
	}
}

func TestValidate_EnvLiveOverridesRefuse(t *testing.T) {
	t.Setenv("AURUMFLOW_API_ENVIRONMENT", "live")
	c := &Config{API: validDemoAPI()}
	c.ApplyEnvOverrides()
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "LIVE") {
		t.Fatalf("env LIVE must refuse, got %v", err)
	}
}

func TestLoad_ExampleTemplateNeedsSecrets(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// tests run from package dir
	p := filepath.Join(root, "example_config.json")
	if _, err := os.Stat(p); err != nil {
		t.Skip(err)
	}
	_, err = Load(p)
	if err == nil {
		t.Fatal("example with empty credentials must not Load without env secrets")
	}
}

func TestParseExecutionMode_DefaultDisabled(t *testing.T) {
	m, err := ParseExecutionMode("")
	if err != nil || m != ExecutionDisabled {
		t.Fatalf("got %s %v", m, err)
	}
}

func TestIsLiveCapitalHost(t *testing.T) {
	if !IsLiveCapitalHost(LiveAPIHost) {
		t.Fatal("live host")
	}
	if IsLiveCapitalHost(DemoAPIHost) {
		t.Fatal("demo must not be live")
	}
	if IsLiveCapitalHost("http://127.0.0.1:1") {
		t.Fatal("httptest is not live")
	}
}

func TestLoad_LiveFileNotConvertedByDemoEnv(t *testing.T) {
	t.Setenv("AURUMFLOW_API_ENVIRONMENT", "demo")
	t.Setenv("AURUMFLOW_EXECUTION_MODE", "DISABLED")
	t.Setenv("AURUMFLOW_API_KEY", "k")
	t.Setenv("AURUMFLOW_IDENTIFIER", "id")
	t.Setenv("AURUMFLOW_PASSWORD", "pw")
	p := filepath.Join(t.TempDir(), "live.json")
	body := `{"api":{"environment":"live","mode":"live","api_key":"k","identifier":"id","password":"pw"},"risk":{"risk_per_trade":0.5}}`
	if err := os.WriteFile(p, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(p); err == nil || !strings.Contains(err.Error(), "LIVE") {
		t.Fatalf("LIVE file must refuse even if env says demo, got %v", err)
	}
}

func TestLoad_LiveURLInFileRefused(t *testing.T) {
	t.Setenv("AURUMFLOW_API_KEY", "k")
	t.Setenv("AURUMFLOW_IDENTIFIER", "id")
	t.Setenv("AURUMFLOW_PASSWORD", "pw")
	p := filepath.Join(t.TempDir(), "url.json")
	body := `{"api":{"environment":"demo","api_base_url":"https://api-capital.backend-capital.com","api_key":"k","identifier":"id","password":"pw"},"risk":{"risk_per_trade":0.5}}`
	if err := os.WriteFile(p, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(p); err == nil || !strings.Contains(err.Error(), "LIVE") {
		t.Fatalf("LIVE URL in file must refuse, got %v", err)
	}
}

func TestRedactAccountID(t *testing.T) {
	if RedactAccountID("ABCDEFGH") != "AB****GH" {
		t.Fatalf("got %s", RedactAccountID("ABCDEFGH"))
	}
}
