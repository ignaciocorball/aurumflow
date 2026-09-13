package telemetry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// resolveCredentials returns service account JSON bytes and path for file-based option.
// Priority: (1) FIREBASE_SERVICE_ACCOUNT_JSON env, (2) GOOGLE_APPLICATION_CREDENTIALS env, (3) config path.
// Returns (jsonBytes, "", nil) when using JSON env; ("", path, nil) when using file; (nil, "", err) on failure.
func resolveCredentials(serviceAccountPath string) (jsonBytes []byte, filePath string, err error) {
	// 1) ENV JSON string
	if s := strings.TrimSpace(os.Getenv("FIREBASE_SERVICE_ACCOUNT_JSON")); s != "" {
		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(s), &raw); err != nil {
			return nil, "", err
		}
		return []byte(s), "", nil
	}
	// 2) ENV path
	if p := strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")); p != "" {
		path := p
		if !filepath.IsAbs(path) {
			cwd, _ := os.Getwd()
			path = filepath.Join(cwd, path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, "", err
		}
		return data, "", nil
	}
	// 3) Config path
	if p := strings.TrimSpace(serviceAccountPath); p != "" {
		path := p
		if !filepath.IsAbs(path) {
			cwd, _ := os.Getwd()
			path = filepath.Join(cwd, path)
		}
		if _, err := os.Stat(path); err != nil {
			return nil, "", err
		}
		return nil, path, nil
	}
	return nil, "", nil
}

// validateCredentials checks that the JSON has project_id, client_email, private_key.
func validateCredentials(data []byte) bool {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return false
	}
	_, hasProject := m["project_id"]
	_, hasEmail := m["client_email"]
	_, hasKey := m["private_key"]
	return hasProject && hasEmail && hasKey
}
