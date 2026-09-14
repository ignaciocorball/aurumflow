package terminal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ReadClient struct {
	http *http.Client
}

func NewReadClient(timeout time.Duration) *ReadClient {
	if timeout <= 0 {
		timeout = 4 * time.Second
	}
	return &ReadClient{http: &http.Client{
		Timeout: timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}}
}

func (c *ReadClient) GetJSON(url string, dest any) error {
	raw, _, err := c.GetBytes(url, "", "")
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dest)
}

func (c *ReadClient) GetBytes(url, etag, modified string) (body []byte, hdr http.Header, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if !AllowHTTP(req.Method) {
		return nil, nil, fmt.Errorf("method denied")
	}
	req.Header.Set("User-Agent", "AurumFlow-Terminal/1.0 (read-only observatory)")
	req.Header.Set("Accept", "application/json, application/rss+xml, application/atom+xml, application/xml, text/xml, */*")
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	if modified != "" {
		req.Header.Set("If-Modified-Since", modified)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, resp.Header, err
	}
	if resp.StatusCode == http.StatusNotModified {
		return nil, resp.Header, errNotModified
	}
	if resp.StatusCode >= 400 {
		return nil, resp.Header, fmt.Errorf("%s: %s", url, resp.Status)
	}
	return b, resp.Header, nil
}

type notModified struct{}

func (notModified) Error() string { return "not modified" }

var errNotModified = notModified{}

func IsNotModified(err error) bool {
	return err == errNotModified
}

func joinURL(base, path string) string {
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(path, "/")
}
