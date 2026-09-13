package secctx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"aurumflow/internal/md"
)

const UserAgent = "AurumFlowResearch/1.0 (research@localhost)"

type Company struct {
	Ticker, CIK, Title string
}

type Filing struct {
	Accession   string
	Form        string
	FiledAt     time.Time
	ReportPeriod time.Time
	AvailableAt time.Time
	Source      string
}

type Fact struct {
	Concept     string
	Value       float64
	FiledAt     time.Time
	PeriodEnd   time.Time
	AvailableAt time.Time
	Source      string
}

type Edge struct {
	From, To, Rel string
	Effective     string
	FiledAt       time.Time
	AvailableAt   time.Time
	Source        string
	Confidence    float64
}

type Graph struct {
	Companies []Company
	Filings   []Filing
	Facts     []Fact
	Edges     []Edge
}

type Provider struct {
	HTTP *http.Client
	Base string
	Dir  string
}

func New(dir string) *Provider {
	return &Provider{
		HTTP: &http.Client{Timeout: 30 * time.Second},
		Base: "https://data.sec.gov",
		Dir:  dir,
	}
}

func (p *Provider) get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/json")
	resp, err := p.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("SEC HTTP %d %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

func (p *Provider) Tickers(ctx context.Context) (map[string]Company, error) {
	raw, err := p.get(ctx, "https://www.sec.gov/files/company_tickers.json")
	if err != nil {
		return nil, err
	}
	var wrap map[string]struct {
		CIK    int    `json:"cik_str"`
		Ticker string `json:"ticker"`
		Title  string `json:"title"`
	}
	if err := json.Unmarshal(raw, &wrap); err != nil {
		return nil, err
	}
	out := map[string]Company{}
	for _, v := range wrap {
		cik := fmt.Sprintf("%010d", v.CIK)
		out[strings.ToUpper(v.Ticker)] = Company{Ticker: strings.ToUpper(v.Ticker), CIK: cik, Title: v.Title}
	}
	return out, nil
}

func (p *Provider) Submissions(ctx context.Context, cik string) ([]Filing, error) {
	url := fmt.Sprintf("%s/submissions/CIK%s.json", strings.TrimRight(p.Base, "/"), cik)
	raw, err := p.get(ctx, url)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Filings struct {
			Recent struct {
				Accession []string `json:"accessionNumber"`
				Form      []string `json:"form"`
				Filed     []string `json:"filingDate"`
				Report    []string `json:"reportDate"`
			} `json:"recent"`
		} `json:"filings"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	var out []Filing
	for i := range doc.Filings.Recent.Accession {
		if i > 40 {
			break
		}
		filed, _ := time.Parse("2006-01-02", doc.Filings.Recent.Filed[i])
		rep := filed
		if i < len(doc.Filings.Recent.Report) && doc.Filings.Recent.Report[i] != "" {
			rep, _ = time.Parse("2006-01-02", doc.Filings.Recent.Report[i])
		}
		avail := filed.Add(24 * time.Hour) // conservative availability after file date
		out = append(out, Filing{
			Accession: doc.Filings.Recent.Accession[i], Form: doc.Filings.Recent.Form[i],
			FiledAt: filed, ReportPeriod: rep, AvailableAt: avail, Source: "data.sec.gov/submissions",
		})
	}
	return out, nil
}

func LatestAvailable(filings []Filing, at time.Time) *Filing {
	var best *Filing
	for i := range filings {
		f := &filings[i]
		if !f.AvailableAt.After(at) {
			if best == nil || f.AvailableAt.After(best.AvailableAt) {
				best = f
			}
		}
	}
	return best
}

func (p *Provider) CompanyFacts(ctx context.Context, cik string) ([]Fact, error) {
	url := fmt.Sprintf("%s/api/xbrl/companyfacts/CIK%s.json", strings.TrimRight(p.Base, "/"), cik)
	raw, err := p.get(ctx, url)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Facts map[string]map[string]struct {
			Units map[string][]struct {
				Val    float64 `json:"val"`
				End    string  `json:"end"`
				Filed  string  `json:"filed"`
				Form   string  `json:"form"`
				Accn   string  `json:"accn"`
			} `json:"units"`
		} `json:"facts"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	want := map[string]bool{"Assets": true, "Revenues": true, "NetIncomeLoss": true, "StockholdersEquity": true}
	var out []Fact
	for _, ns := range doc.Facts {
		for concept, body := range ns {
			if !want[concept] {
				continue
			}
			for unit, pts := range body.Units {
				if unit != "USD" && unit != "USD/shares" {
					continue
				}
				start := 0
				if len(pts) > 8 {
					start = len(pts) - 8
				}
				for _, pt := range pts[start:] {
					filed, _ := time.Parse("2006-01-02", pt.Filed)
					end, _ := time.Parse("2006-01-02", pt.End)
					out = append(out, Fact{
						Concept: concept, Value: pt.Val, FiledAt: filed, PeriodEnd: end,
						AvailableAt: filed.Add(24 * time.Hour), Source: "data.sec.gov/api/xbrl/companyfacts",
					})
				}
			}
		}
	}
	return out, nil
}

func (g *Graph) AddCompanyFilings(c Company, filings []Filing) {
	g.Companies = append(g.Companies, c)
	for _, f := range filings {
		g.Filings = append(g.Filings, f)
		g.Edges = append(g.Edges, Edge{
			From: c.CIK, To: f.Accession, Rel: "company_filing",
			Effective: f.ReportPeriod.Format("2006-01-02"), FiledAt: f.FiledAt, AvailableAt: f.AvailableAt,
			Source: f.Source, Confidence: 1,
		})
		if strings.HasPrefix(f.Form, "13F") {
			g.Edges = append(g.Edges, Edge{
				From: "manager:" + f.Accession, To: c.CIK, Rel: "manager_holding_placeholder",
				Effective: f.ReportPeriod.Format("2006-01-02"), FiledAt: f.FiledAt, AvailableAt: f.AvailableAt,
				Source: f.Source, Confidence: 0.4,
			})
		}
		if f.Form == "4" || f.Form == "3" || f.Form == "5" {
			g.Edges = append(g.Edges, Edge{
				From: "insider:" + f.Accession, To: c.CIK, Rel: "insider_company",
				Effective: f.ReportPeriod.Format("2006-01-02"), FiledAt: f.FiledAt, AvailableAt: f.AvailableAt,
				Source: f.Source, Confidence: 0.6,
			})
		}
	}
}

func (g *Graph) AddFacts(c Company, facts []Fact) {
	for _, f := range facts {
		g.Facts = append(g.Facts, f)
		g.Edges = append(g.Edges, Edge{
			From: c.CIK, To: "fact:" + f.Concept + ":" + f.PeriodEnd.Format("2006-01-02"), Rel: "company_fact",
			Effective: f.PeriodEnd.Format("2006-01-02"), FiledAt: f.FiledAt, AvailableAt: f.AvailableAt,
			Source: f.Source, Confidence: 1,
		})
	}
}

func (g *Graph) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func ContextAt(filings []Filing, at time.Time) md.Provenance {
	f := LatestAvailable(filings, at)
	p := md.Provenance{Provider: "sec", Venue: "edgar", Quality: md.QualityDelayed, FreshnessClass: md.FreshSlow, IsDelayed: true}
	if f != nil {
		p.EventTime = f.ReportPeriod
		p.PublishedAt = md.PointerTime(f.FiledAt)
		p.AvailableAt = f.AvailableAt
	}
	return p
}

func ParseCIK(s string) string {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return fmt.Sprintf("%010d", n)
}
