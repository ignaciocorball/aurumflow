package catalog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Dataset struct {
	Provider      string `json:"provider"`
	Instrument    string `json:"instrument"`
	Type          string `json:"type"`
	From          string `json:"from"`
	To            string `json:"to"`
	Rows          int    `json:"rows"`
	Quality       string `json:"quality"`
	Gaps          int    `json:"gaps"`
	Checksum      string `json:"checksum"`
	DownloadedAt  string `json:"downloaded_at"`
	SourceVersion string `json:"source_version"`
	Path          string `json:"path,omitempty"`
	Status        string `json:"status,omitempty"`
}

type Catalog struct {
	mu       sync.Mutex
	Datasets []Dataset `json:"datasets"`
}

func Load(path string) (*Catalog, error) {
	c := &Catalog{}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Catalog) Upsert(d Dataset) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := d.Provider + "|" + d.Instrument + "|" + d.Type + "|" + d.From + "|" + d.To
	for i, x := range c.Datasets {
		k := x.Provider + "|" + x.Instrument + "|" + x.Type + "|" + x.From + "|" + x.To
		if k == key {
			c.Datasets[i] = d
			return
		}
	}
	c.Datasets = append(c.Datasets, d)
}

func (c *Catalog) Save(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func FileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func NowUTC() string { return time.Now().UTC().Format(time.RFC3339) }
