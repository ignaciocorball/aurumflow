package news

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
)

type Item struct {
	ID          string    `json:"id"`
	Source      string    `json:"source"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"published_at"`
	IngestedAt  time.Time `json:"ingested_at"`
	Summary     string    `json:"summary"`
	Markets     []string  `json:"markets"`
	Themes      []string  `json:"themes"`
	Region      string    `json:"region"`
	ClusterID   string    `json:"cluster_id"`
}

type Cluster struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Summary    string    `json:"summary"`
	Source     string    `json:"source"`
	Sources    []string  `json:"sources"`
	URL        string    `json:"url"`
	Published  time.Time `json:"published"`
	Markets    []string  `json:"markets"`
	Themes     []string  `json:"themes"`
	Region     string    `json:"region"`
	Relevance  float64   `json:"relevance"`
	Importance string    `json:"importance"`
	Count      int       `json:"count"`
}

type Source struct {
	ID      string
	Name    string
	URL     string
	Cadence time.Duration
	Class   string
}

func DefaultSources() []Source {
	return []Source{
		{ID: "fed", Name: "Federal Reserve", URL: "https://www.federalreserve.gov/feeds/press_all.xml", Cadence: 5 * time.Minute, Class: "CRITICAL"},
		{ID: "ecb", Name: "European Central Bank", URL: "https://www.ecb.europa.eu/rss/press.html", Cadence: 15 * time.Minute, Class: "STANDARD"},
		{ID: "boe", Name: "Bank of England", URL: "https://www.bankofengland.co.uk/rss/news", Cadence: 15 * time.Minute, Class: "STANDARD"},
		{ID: "bls", Name: "U.S. Bureau of Labor Statistics", URL: "https://www.bls.gov/feed/bls_latest.rss", Cadence: 30 * time.Minute, Class: "SLOW"},
		{ID: "eia", Name: "U.S. Energy Information Administration", URL: "https://www.eia.gov/rss/todayinenergy.xml", Cadence: 30 * time.Minute, Class: "SLOW"},
	}
}

type Store struct {
	mu    sync.Mutex
	Items []Item `json:"items"`
	path  string
}

func NewStore(path string) *Store {
	s := &Store{path: path}
	_ = s.load()
	return s
}

func (s *Store) load() error {
	b, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, s)
}

func (s *Store) persist() {
	if s.path == "" {
		return
	}
	_ = os.MkdirAll(filepath.Dir(s.path), 0o755)
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(s.path, b, 0o644)
}

func (s *Store) Ingest(items []Item) (added int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	seen := map[string]bool{}
	for _, it := range s.Items {
		seen[it.ID] = true
		if it.URL != "" {
			seen["u:"+canonURL(it.URL)] = true
		}
	}
	now := time.Now().UTC()
	for _, it := range items {
		it.Title = strings.TrimSpace(it.Title)
		it.URL = strings.TrimSpace(it.URL)
		it.Summary = clip(stripTags(it.Summary), 280)
		if it.Title == "" {
			continue
		}
		if it.ID == "" {
			it.ID = hashID(it.Source + "|" + canonURL(it.URL) + "|" + normalizeTitle(it.Title))
		}
		if seen[it.ID] || (it.URL != "" && seen["u:"+canonURL(it.URL)]) {
			continue
		}
		if dupTitle(s.Items, it) {
			continue
		}
		it.IngestedAt = now
		it.Markets, it.Themes, it.Region = Classify(it.Title + " " + it.Summary)
		s.Items = append(s.Items, it)
		seen[it.ID] = true
		added++
	}
	if len(s.Items) > 400 {
		sort.Slice(s.Items, func(i, j int) bool { return s.Items[i].PublishedAt.After(s.Items[j].PublishedAt) })
		s.Items = s.Items[:400]
	}
	if added > 0 {
		s.persist()
	}
	return added
}

func (s *Store) Clusters(now time.Time) []Cluster {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := append([]Item(nil), s.Items...)
	return ClusterItems(items, now)
}

func dupTitle(have []Item, it Item) bool {
	nt := normalizeTitle(it.Title)
	for _, h := range have {
		if h.Source == it.Source && normalizeTitle(h.Title) == nt {
			return true
		}
	}
	return false
}

func ClusterItems(items []Item, now time.Time) []Cluster {
	sort.Slice(items, func(i, j int) bool { return items[i].PublishedAt.After(items[j].PublishedAt) })
	used := make([]bool, len(items))
	var out []Cluster
	for i := range items {
		if used[i] {
			continue
		}
		group := []Item{items[i]}
		used[i] = true
		for j := i + 1; j < len(items); j++ {
			if used[j] {
				continue
			}
			if similar(normalizeTitle(items[i].Title), normalizeTitle(items[j].Title)) >= 0.72 && hoursApart(items[i].PublishedAt, items[j].PublishedAt) < 48 {
				group = append(group, items[j])
				used[j] = true
			}
		}
		c := clusterFrom(group, now)
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Relevance > out[j].Relevance })
	if len(out) > 80 {
		out = out[:80]
	}
	return out
}

func clusterFrom(group []Item, now time.Time) Cluster {
	head := group[0]
	srcset := []string{}
	seen := map[string]bool{}
	markets := map[string]bool{}
	themes := map[string]bool{}
	for _, it := range group {
		if !seen[it.Source] {
			srcset = append(srcset, it.Source)
			seen[it.Source] = true
		}
		for _, m := range it.Markets {
			markets[m] = true
		}
		for _, t := range it.Themes {
			themes[t] = true
		}
	}
	c := Cluster{
		ID: hashID("c|" + head.ID), Title: head.Title, Summary: head.Summary,
		Source: head.Source, Sources: srcset, URL: head.URL, Published: head.PublishedAt,
		Region: head.Region, Count: len(group),
	}
	for m := range markets {
		c.Markets = append(c.Markets, m)
	}
	for t := range themes {
		c.Themes = append(c.Themes, t)
	}
	sort.Strings(c.Markets)
	sort.Strings(c.Themes)
	c.Relevance = Relevance(c, now)
	c.Importance = importanceOf(c)
	return c
}

func Relevance(c Cluster, now time.Time) float64 {
	score := 0.0
	official := map[string]bool{"Federal Reserve": true, "European Central Bank": true, "Bank of England": true, "U.S. Bureau of Labor Statistics": true, "U.S. Energy Information Administration": true}
	if official[c.Source] {
		score += 40
	}
	age := now.Sub(c.Published)
	if age < 6*time.Hour {
		score += 30
	} else if age < 24*time.Hour {
		score += 15
	} else if age < 72*time.Hour {
		score += 5
	}
	m := len(c.Markets)
	if m > 3 {
		m = 3
	}
	score += 10 * float64(m)
	extra := c.Count - 1
	if extra > 4 {
		extra = 4
	}
	score += 5 * float64(extra)
	if score > 100 {
		score = 100
	}
	return score
}

func importanceOf(c Cluster) string {
	if c.Relevance >= 70 || contains(c.Themes, "MONETARY_POLICY") || contains(c.Themes, "LABOR") {
		return "HIGH"
	}
	if c.Relevance >= 45 {
		return "MEDIUM"
	}
	return "LOW"
}

func contains(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

func Classify(text string) (markets []string, themes []string, region string) {
	u := strings.ToUpper(text)
	add := func(dst *[]string, v string) {
		for _, x := range *dst {
			if x == v {
				return
			}
		}
		*dst = append(*dst, v)
	}
	if strings.Contains(u, "GOLD") || strings.Contains(u, "PRECIOUS") {
		add(&markets, "GOLD")
	}
	if strings.Contains(u, "SILVER") {
		add(&markets, "SILVER")
	}
	if strings.Contains(u, "OIL") || strings.Contains(u, "PETROLEUM") || strings.Contains(u, "CRUDE") || strings.Contains(u, "WTI") || strings.Contains(u, "BRENT") {
		add(&markets, "OIL")
	}
	if strings.Contains(u, "NASDAQ") || strings.Contains(u, "TECHNOLOGY") || strings.Contains(u, "US100") {
		add(&markets, "US100")
	}
	if strings.Contains(u, "S&P") || strings.Contains(u, "S & P") || strings.Contains(u, "US500") {
		add(&markets, "US500")
	}
	if strings.Contains(u, "DOW") || strings.Contains(u, "US30") {
		add(&markets, "US30")
	}
	if strings.Contains(u, "NIKKEI") || strings.Contains(u, "JAPAN") || strings.Contains(u, "BOJ") {
		add(&markets, "J225")
		region = "JAPAN"
	}
	if strings.Contains(u, "DAX") || strings.Contains(u, "EURO") || strings.Contains(u, "ECB") {
		add(&markets, "DE40")
		if region == "" {
			region = "EUROPE"
		}
	}
	if strings.Contains(u, "FTSE") || strings.Contains(u, "BOE") || strings.Contains(u, "BANK OF ENGLAND") {
		add(&markets, "UK100")
		if region == "" {
			region = "UNITED_KINGDOM"
		}
	}
	if strings.Contains(u, "CHINA") || strings.Contains(u, "HONG KONG") || strings.Contains(u, "PBOC") {
		add(&markets, "CN50")
		if region == "" {
			region = "CHINA_HONG_KONG"
		}
	}
	if strings.Contains(u, "BITCOIN") || strings.Contains(u, "BTC") || strings.Contains(u, "CRYPTO") {
		add(&markets, "BTC")
	}
	if strings.Contains(u, "FOMC") || strings.Contains(u, "FEDERAL RESERVE") || strings.Contains(u, "INTEREST RATE") || strings.Contains(u, "MONETARY POLICY") {
		add(&themes, "MONETARY_POLICY")
		add(&markets, "US100")
		add(&markets, "US500")
		add(&markets, "GOLD")
		if region == "" {
			region = "UNITED_STATES"
		}
	}
	if strings.Contains(u, "CPI") || strings.Contains(u, "INFLATION") || strings.Contains(u, "PCE") {
		add(&themes, "INFLATION")
	}
	if strings.Contains(u, "PAYROLL") || strings.Contains(u, "UNEMPLOYMENT") || strings.Contains(u, "NFP") || strings.Contains(u, "EMPLOYMENT") {
		add(&themes, "LABOR")
	}
	if strings.Contains(u, "GDP") {
		add(&themes, "GROWTH")
	}
	if strings.Contains(u, "OIL") || strings.Contains(u, "ENERGY") {
		add(&themes, "ENERGY")
	}
	if strings.Contains(u, "SANCTION") || strings.Contains(u, "WAR") || strings.Contains(u, "GEOPOLIT") {
		add(&themes, "GEOPOLITICS")
	}
	if strings.Contains(u, "REGULAT") || strings.Contains(u, "SEC ") {
		add(&themes, "REGULATION")
	}
	if strings.Contains(u, "EARNINGS") {
		add(&themes, "EARNINGS")
	}
	if strings.Contains(u, "LIQUIDITY") || strings.Contains(u, "BALANCE SHEET") {
		add(&themes, "LIQUIDITY")
	}
	if region == "" && (strings.Contains(u, "UNITED STATES") || strings.Contains(u, "U.S.") || strings.Contains(u, "FED")) {
		region = "UNITED_STATES"
	}
	return
}

func ParseFeed(source string, body []byte) []Item {
	var rss rssDoc
	if xml.Unmarshal(body, &rss) == nil && len(rss.Channel.Items) > 0 {
		var out []Item
		for _, it := range rss.Channel.Items {
			out = append(out, Item{
				Source: source, Title: strings.TrimSpace(it.Title),
				URL: firstNonEmpty(it.Link, it.GUID), Summary: firstNonEmpty(it.Description, it.Encoded),
				PublishedAt: parseRSSTime(it.PubDate),
			})
		}
		return out
	}
	var atom atomDoc
	if xml.Unmarshal(body, &atom) == nil && len(atom.Entries) > 0 {
		var out []Item
		for _, e := range atom.Entries {
			link := e.ID
			for _, l := range e.Links {
				if l.Rel == "" || l.Rel == "alternate" {
					link = l.Href
					break
				}
			}
			out = append(out, Item{
				Source: source, Title: strings.TrimSpace(e.Title), URL: link,
				Summary: firstNonEmpty(e.Summary, e.Content),
				PublishedAt: parseRSSTime(firstNonEmpty(e.Published, e.Updated)),
			})
		}
		return out
	}
	return nil
}

type rssDoc struct {
	Channel struct {
		Items []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			GUID        string `xml:"guid"`
			Description string `xml:"description"`
			Encoded     string `xml:"encoded"`
			PubDate     string `xml:"pubDate"`
		} `xml:"item"`
	} `xml:"channel"`
}

type atomDoc struct {
	Entries []struct {
		Title     string `xml:"title"`
		ID        string `xml:"id"`
		Summary   string `xml:"summary"`
		Content   string `xml:"content"`
		Published string `xml:"published"`
		Updated   string `xml:"updated"`
		Links     []struct {
			Rel  string `xml:"rel,attr"`
			Href string `xml:"href,attr"`
		} `xml:"link"`
	} `xml:"entry"`
}

func parseRSSTime(s string) time.Time {
	s = strings.TrimSpace(s)
	fmts := []string{time.RFC1123Z, time.RFC1123, time.RFC3339, time.RFC822Z, time.RFC822, "Mon, 02 Jan 2006 15:04:05 MST"}
	for _, f := range fmts {
		if t, err := time.Parse(f, s); err == nil {
			return t.UTC()
		}
	}
	return time.Now().UTC()
}

func normalizeTitle(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func similar(a, b string) float64 {
	if a == "" || b == "" {
		return 0
	}
	if a == b {
		return 1
	}
	ta := strings.Fields(a)
	tb := strings.Fields(b)
	set := map[string]bool{}
	for _, t := range ta {
		set[t] = true
	}
	inter := 0
	for _, t := range tb {
		if set[t] {
			inter++
		}
	}
	union := len(ta) + len(tb) - inter
	if union == 0 {
		return 0
	}
	return float64(inter) / float64(union)
}

func hoursApart(a, b time.Time) float64 {
	if a.IsZero() || b.IsZero() {
		return 0
	}
	d := a.Sub(b)
	if d < 0 {
		d = -d
	}
	return d.Hours()
}

func canonURL(u string) string {
	u = strings.TrimSpace(u)
	u = strings.TrimSuffix(u, "/")
	if i := strings.Index(u, "?"); i >= 0 {
		u = u[:i]
	}
	return strings.ToLower(u)
}

func hashID(s string) string {
	h := sha1.Sum([]byte(s))
	return hex.EncodeToString(h[:10])
}

func firstNonEmpty(xs ...string) string {
	for _, x := range xs {
		if strings.TrimSpace(x) != "" {
			return strings.TrimSpace(x)
		}
	}
	return ""
}

var tagRe = regexp.MustCompile(`<[^>]+>`)

func stripTags(s string) string {
	return strings.TrimSpace(tagRe.ReplaceAllString(s, " "))
}

func clip(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
