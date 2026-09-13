package globalsources

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type HTTPCache struct {
	mu       sync.Mutex
	Client   *http.Client
	Dir      string
	etag     map[string]string
	lastMod  map[string]string
	lastHit  map[string]time.Time
	MinGap   time.Duration
	Backoff  time.Duration
}

func NewHTTPCache(dir string) *HTTPCache {
	return &HTTPCache{
		Client:  &http.Client{Timeout: 90 * time.Second},
		Dir:     dir,
		etag:    map[string]string{},
		lastMod: map[string]string{},
		lastHit: map[string]time.Time{},
		MinGap:  2 * time.Second,
		Backoff: 5 * time.Second,
	}
}

func (h *HTTPCache) GetFresh(ctx context.Context, url, cacheName string) ([]byte, string, error) {
	return h.get(ctx, url, cacheName, true)
}

func (h *HTTPCache) Get(ctx context.Context, url, cacheName string) ([]byte, string, error) {
	return h.get(ctx, url, cacheName, false)
}

func (h *HTTPCache) get(ctx context.Context, url, cacheName string, fresh bool) ([]byte, string, error) {
	cachedPath := ""
	if h.Dir != "" && cacheName != "" {
		cachedPath = filepath.Join(h.Dir, cacheName)
		if !fresh {
			if raw, err := os.ReadFile(cachedPath); err == nil && len(raw) > 0 {
				return raw, "cache", nil
			}
		}
	}
	h.mu.Lock()
	if last, ok := h.lastHit[url]; ok && h.MinGap > 0 && time.Since(last) < h.MinGap {
		h.mu.Unlock()
		time.Sleep(h.Backoff)
		h.mu.Lock()
	}
	h.lastHit[url] = time.Now()
	et, lm := h.etag[url], h.lastMod[url]
	h.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; AurumFlowResearch/1.0; +https://localhost)")
	if et != "" {
		req.Header.Set("If-None-Match", et)
	}
	if lm != "" {
		req.Header.Set("If-Modified-Since", lm)
	}
	resp, err := h.Client.Do(req)
	if err != nil {
		if cachedPath != "" {
			if raw, rerr := os.ReadFile(cachedPath); rerr == nil && len(raw) > 0 {
				return raw, "cache", err
			}
		}
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotModified && cachedPath != "" {
		if raw, err := os.ReadFile(cachedPath); err == nil {
			return raw, "not_modified", nil
		}
	}
	if resp.StatusCode != http.StatusOK {
		if cachedPath != "" {
			if raw, rerr := os.ReadFile(cachedPath); rerr == nil && len(raw) > 0 {
				return raw, "cache", errStatus(resp.StatusCode)
			}
		}
		return nil, "", errStatus(resp.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, "", err
	}
	h.mu.Lock()
	if v := resp.Header.Get("ETag"); v != "" {
		h.etag[url] = v
	}
	if v := resp.Header.Get("Last-Modified"); v != "" {
		h.lastMod[url] = v
	}
	h.mu.Unlock()
	if h.Dir != "" && cacheName != "" {
		_ = os.MkdirAll(h.Dir, 0o755)
		_ = os.WriteFile(filepath.Join(h.Dir, cacheName), raw, 0o644)
	}
	return raw, "network", nil
}

type statusErr int

func errStatus(c int) error { return statusErr(c) }
func (e statusErr) Error() string { return "http " + itoa(int(e)) }

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [16]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
