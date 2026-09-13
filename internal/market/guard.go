package market

import (
	"fmt"
	"strings"

	"aurumflow/config"
)

// ErrLiveBrokerDisabled is returned if any request would target Capital.com LIVE.
var ErrLiveBrokerDisabled = fmt.Errorf("%s", config.ErrLiveDisabled)

// IsLiveCapitalHost reports whether baseURL is the Capital.com LIVE host.
func IsLiveCapitalHost(baseURL string) bool {
	return config.IsLiveCapitalHost(baseURL)
}

func (c *Client) assertNotLive(path string) error {
	if c == nil {
		return fmt.Errorf("market client is nil")
	}
	if IsLiveCapitalHost(c.BaseURL) {
		return fmt.Errorf("%w (refused %s %s)", ErrLiveBrokerDisabled, strings.TrimSpace(path), c.BaseURL)
	}
	return nil
}
