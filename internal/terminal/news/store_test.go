package news

import (
	"testing"
	"time"
)

func TestDedupAndCluster(t *testing.T) {
	now := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	items := []Item{
		{Source: "Federal Reserve", Title: "Federal Reserve issues FOMC statement", URL: "https://federalreserve.gov/a", PublishedAt: now, Summary: "Rates unchanged"},
		{Source: "Federal Reserve", Title: "Federal Reserve issues FOMC statement", URL: "https://federalreserve.gov/a/", PublishedAt: now},
		{Source: "European Central Bank", Title: "Federal Reserve issues FOMC statement", URL: "https://ecb.europa.eu/x", PublishedAt: now.Add(-time.Minute), Summary: "echo"},
	}
	s := NewStore("")
	if s.Ingest(items) < 1 {
		t.Fatal("ingest")
	}
	if s.Ingest(items) != 0 {
		t.Fatal("second ingest should dedupe")
	}
	cs := ClusterItems(s.Items, now)
	if len(cs) != 1 {
		t.Fatalf("clusters %d", len(cs))
	}
	if cs[0].Count < 2 {
		t.Fatalf("count %d sources %v", cs[0].Count, cs[0].Sources)
	}
	if cs[0].Relevance <= 0 {
		t.Fatal("relevance")
	}
}

func TestClassifyMarkets(t *testing.T) {
	m, themes, region := Classify("FOMC holds rates; gold and Nasdaq futures react")
	if region != "UNITED_STATES" {
		t.Fatal(region)
	}
	has := map[string]bool{}
	for _, x := range m {
		has[x] = true
	}
	if !has["GOLD"] || !has["US100"] {
		t.Fatalf("%v", m)
	}
	th := map[string]bool{}
	for _, x := range themes {
		th[x] = true
	}
	if !th["MONETARY_POLICY"] {
		t.Fatalf("%v", themes)
	}
}

func TestParseRSS(t *testing.T) {
	body := []byte(`<?xml version="1.0"?><rss><channel><item><title>Oil inventory falls</title><link>https://eia.gov/1</link><pubDate>Mon, 14 Sep 2026 10:00:00 GMT</pubDate><description>Crude stocks</description></item></channel></rss>`)
	items := ParseFeed("U.S. Energy Information Administration", body)
	if len(items) != 1 || items[0].Title != "Oil inventory falls" {
		t.Fatalf("%+v", items)
	}
}
