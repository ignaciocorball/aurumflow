package telemetry

import (
	"path/filepath"
	"strings"
)

// ExtractBotTag extracts the bot directory name from config path.
// Returns empty string if config is in config/ or root.
// Examples:
//   - "bots/eth-offensive-alpha/config.json" → "eth-offensive-alpha"
//   - "config/config.json" → ""
func ExtractBotTag(configPath string) string {
	dir := filepath.Dir(configPath)
	base := filepath.Base(dir)
	// If we're in config/ or root, return empty
	if base == "config" || base == "." || base == "" {
		return ""
	}
	// On Windows, check if base is a drive letter (C:, D:, etc.)
	if len(base) == 2 && base[1] == ':' {
		return ""
	}
	return base
}

// NormalizeBotTag normalizes a bot tag to lowercase with spaces instead of hyphens.
// Removes underscores, keeps hyphens converted to spaces, removes special characters.
// Max length: 50 characters.
// Examples:
//   - "eth-offensive-alpha" → "eth offensive alpha"
//   - "eth-defensive-core" → "eth defensive core"
//   - "xauusd-live-01" → "xauusd live 01"
func NormalizeBotTag(tag string) string {
	s := strings.ToLower(strings.TrimSpace(tag))
	if s == "" {
		return ""
	}
	// Replace hyphens and underscores with spaces
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	// Collapse multiple spaces to single space
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	// Keep only alphanumeric and spaces
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' {
			b.WriteRune(r)
		}
	}
	result := strings.TrimSpace(b.String())
	if len(result) > 50 {
		result = result[:50]
		result = strings.TrimSpace(result)
	}
	return result
}
