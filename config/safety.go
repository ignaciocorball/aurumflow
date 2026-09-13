package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// APIEnvironment is the broker environment. P1 accepts only DEMO.
type APIEnvironment string

const (
	APIEnvDemo APIEnvironment = "demo"
	APIEnvLive APIEnvironment = "live" // parseable; always fatal in P1
)

// ExecutionMode is permission to mutate the broker. LIVE is not a supported P1 mode.
type ExecutionMode string

const (
	ExecutionDisabled ExecutionMode = "DISABLED"
	ExecutionDryRun   ExecutionMode = "DRY_RUN"
	ExecutionDemo     ExecutionMode = "DEMO"
	ExecutionLive     ExecutionMode = "LIVE" // parseable; always fatal in P1
)

const (
	// DemoAPIHost is the only Capital.com host permitted in P1 operational runtime.
	DemoAPIHost = "https://demo-api-capital.backend-capital.com"
	// LiveAPIHost is recognized only to refuse it.
	LiveAPIHost = "https://api-capital.backend-capital.com"

	// ErrLiveDisabled is the canonical P1 refusal for any LIVE broker path.
	ErrLiveDisabled = "STARTUP REFUSED: LIVE broker environment is disabled in P1."
)

// IsLiveCapitalHost reports whether u is the Capital.com LIVE API host.
func IsLiveCapitalHost(u string) bool {
	s := strings.ToLower(strings.TrimSpace(u))
	if s == "" {
		return false
	}
	if strings.Contains(s, "demo-api") {
		return false
	}
	return strings.Contains(s, "api-capital.backend-capital.com") ||
		strings.Contains(s, "api-capital.capital.com")
}

// IsDemoCapitalHost reports whether u is the known Capital.com DEMO host.
func IsDemoCapitalHost(u string) bool {
	s := strings.ToLower(strings.TrimSpace(u))
	if s == "" {
		return false
	}
	return strings.Contains(s, "demo-api-capital.backend-capital.com")
}

func isLiveToken(s string) bool {
	return strings.EqualFold(strings.TrimSpace(s), string(APIEnvLive))
}

func isDemoToken(s string) bool {
	return strings.EqualFold(strings.TrimSpace(s), string(APIEnvDemo))
}

// ResolveAPIEnvironment returns the explicit broker environment or an error (fail-closed).
// Empty / missing / invalid / LIVE all refuse. Does not infer DEMO from a URL.
func ResolveAPIEnvironment(environment, mode string) (APIEnvironment, error) {
	raw := strings.TrimSpace(environment)
	if raw == "" {
		raw = strings.TrimSpace(mode)
	}
	if raw == "" {
		return "", fmt.Errorf("config: api.environment is required (demo); missing environment is fail-closed")
	}
	switch strings.ToLower(raw) {
	case string(APIEnvDemo):
		return APIEnvDemo, nil
	case string(APIEnvLive):
		return "", fmt.Errorf("%s", ErrLiveDisabled)
	default:
		return "", fmt.Errorf("config: api.environment %q is invalid; P1 accepts only %q", raw, APIEnvDemo)
	}
}

// ParseExecutionMode parses an execution mode. Empty defaults to DISABLED.
// LIVE and unknown values refuse.
func ParseExecutionMode(raw string) (ExecutionMode, error) {
	s := strings.ToUpper(strings.TrimSpace(raw))
	if s == "" {
		return ExecutionDisabled, nil
	}
	switch ExecutionMode(s) {
	case ExecutionDisabled, ExecutionDryRun, ExecutionDemo:
		return ExecutionMode(s), nil
	case ExecutionLive:
		return "", fmt.Errorf("%s", ErrLiveDisabled)
	default:
		return "", fmt.Errorf("config: api.execution_mode %q is invalid; P1 accepts DISABLED, DRY_RUN, DEMO", raw)
	}
}

// ValidateOperationalHost refuses LIVE and unknown Capital hosts.
// Empty URL is allowed (caller then assigns DemoAPIHost).
// httptest / localhost is not accepted on the operational config path.
func ValidateOperationalHost(raw string) error {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil
	}
	if IsLiveCapitalHost(s) {
		return fmt.Errorf("%s", ErrLiveDisabled)
	}
	if IsDemoCapitalHost(s) {
		return nil
	}
	parsed, err := url.Parse(s)
	if err != nil || parsed.Host == "" {
		return fmt.Errorf("config: api.api_base_url is malformed")
	}
	return fmt.Errorf("config: api.api_base_url is not a permitted P1 host (operational runtime uses the built-in DEMO host only)")
}

// refuseIfConfigDeclaresLive rejects on-disk LIVE environment, mode, host, or execution mode
// before environment variables can rewrite the file into a DEMO-looking config.
func refuseIfConfigDeclaresLive(c *Config) error {
	if c == nil {
		return fmt.Errorf("config: nil")
	}
	if isLiveToken(c.API.Environment) || isLiveToken(c.API.Mode) || IsLiveCapitalHost(c.API.BaseURL) {
		return fmt.Errorf("%s", ErrLiveDisabled)
	}
	if strings.EqualFold(strings.TrimSpace(c.API.ExecutionMode), string(ExecutionLive)) {
		return fmt.Errorf("%s", ErrLiveDisabled)
	}
	return nil
}

// ApplyP1BrokerLock forces DEMO host and rejects any LIVE signal.
func (c *Config) ApplyP1BrokerLock() error {
	if isLiveToken(c.API.Environment) || isLiveToken(c.API.Mode) {
		return fmt.Errorf("%s", ErrLiveDisabled)
	}
	if IsLiveCapitalHost(c.API.BaseURL) {
		return fmt.Errorf("%s", ErrLiveDisabled)
	}
	if v := os.Getenv("AURUMFLOW_API_ENVIRONMENT"); isLiveToken(v) {
		return fmt.Errorf("%s", ErrLiveDisabled)
	}
	if v := os.Getenv("AURUMFLOW_EXECUTION_MODE"); strings.EqualFold(strings.TrimSpace(v), string(ExecutionLive)) {
		return fmt.Errorf("%s", ErrLiveDisabled)
	}

	env, err := ResolveAPIEnvironment(c.API.Environment, c.API.Mode)
	if err != nil {
		return err
	}
	if err := ValidateOperationalHost(c.API.BaseURL); err != nil {
		return err
	}
	mode, err := ParseExecutionMode(c.API.ExecutionMode)
	if err != nil {
		return err
	}

	c.API.Environment = string(env)
	c.API.Mode = string(env)
	c.API.BaseURL = DemoAPIHost
	c.API.ExecutionMode = string(mode)
	return nil
}

// ExecMode returns the parsed execution mode after Validate.
func (c *Config) ExecMode() ExecutionMode {
	m, err := ParseExecutionMode(c.API.ExecutionMode)
	if err != nil {
		return ExecutionDisabled
	}
	return m
}

// RedactAccountID returns a partial account id safe for logs.
func RedactAccountID(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return "(none)"
	}
	if len(id) <= 4 {
		return "****"
	}
	return id[:2] + strings.Repeat("*", len(id)-4) + id[len(id)-2:]
}
