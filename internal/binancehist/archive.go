package binancehist

import (
	"archive/zip"
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"aurumflow/internal/catalog"
	"aurumflow/internal/flow"
	"aurumflow/internal/md"
)

const VisionBase = "https://data.binance.vision"

type AggTrade struct {
	ID           int64
	Price        float64
	Qty          float64
	FirstTradeID int64
	LastTradeID  int64
	Time         time.Time
	BuyerMaker   bool
}

func DailyURL(symbol, day string) string {
	return fmt.Sprintf("%s/data/futures/um/daily/aggTrades/%s/%s-aggTrades-%s.zip", VisionBase, symbol, symbol, day)
}

func DailyChecksumURL(symbol, day string) string {
	return DailyURL(symbol, day) + ".CHECKSUM"
}

func KlineURL(symbol, interval, day string) string {
	return fmt.Sprintf("%s/data/futures/um/daily/klines/%s/%s/%s-%s-%s.zip", VisionBase, symbol, interval, symbol, interval, day)
}

type Archive struct {
	HTTP   *http.Client
	Dir    string
	Symbol string
}

func New(dir, symbol string) *Archive {
	if symbol == "" {
		symbol = "BTCUSDT"
	}
	return &Archive{HTTP: &http.Client{Timeout: 2 * time.Minute}, Dir: dir, Symbol: symbol}
}

func (a *Archive) DayPath(day string) string {
	return filepath.Join(a.Dir, "binance", "aggTrades", a.Symbol, a.Symbol+"-aggTrades-"+day+".zip")
}

func (a *Archive) DownloadDay(ctx context.Context, day string) (string, error) {
	dest := a.DayPath(day)
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}
	if st, err := os.Stat(dest); err == nil && st.Size() > 0 {
		return dest, nil
	}
	part := dest + ".part"
	if err := a.get(ctx, DailyURL(a.Symbol, day), part); err != nil {
		return "", err
	}
	if err := os.Rename(part, dest); err != nil {
		return "", err
	}
	return dest, nil
}

func (a *Archive) VerifyChecksum(ctx context.Context, day, zipPath string) (string, bool, error) {
	sum, err := catalog.FileSHA256(zipPath)
	if err != nil {
		return "", false, err
	}
	tmp := zipPath + ".CHECKSUM"
	if err := a.get(ctx, DailyChecksumURL(a.Symbol, day), tmp); err != nil {
		return sum, false, err
	}
	raw, err := os.ReadFile(tmp)
	if err != nil {
		return sum, false, err
	}
	want := strings.Fields(string(raw))
	if len(want) == 0 {
		return sum, false, fmt.Errorf("empty checksum")
	}
	ok := strings.EqualFold(want[0], sum)
	return sum, ok, nil
}

func ParseChecksumLine(line, got string) bool {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false
	}
	return strings.EqualFold(fields[0], got)
}

func SHA256Bytes(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func (a *Archive) get(ctx context.Context, url, dest string) error {
	var last error
	for i := 0; i < 4; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		resp, err := a.HTTP.Do(req)
		if err != nil {
			last = err
			time.Sleep(time.Duration(1<<i) * 300 * time.Millisecond)
			continue
		}
		if resp.StatusCode != 200 {
			_ = resp.Body.Close()
			last = fmt.Errorf("%s HTTP %d", url, resp.StatusCode)
			if resp.StatusCode == 404 {
				return last
			}
			time.Sleep(time.Duration(1<<i) * 300 * time.Millisecond)
			continue
		}
		f, err := os.Create(dest)
		if err != nil {
			_ = resp.Body.Close()
			return err
		}
		_, err = io.Copy(f, resp.Body)
		_ = resp.Body.Close()
		_ = f.Close()
		if err != nil {
			last = err
			continue
		}
		return nil
	}
	return last
}

func IterZipCSV(zipPath string, fn func(row []string) error) error {
	return iterZipCSV(zipPath, true, fn)
}

func IterTrades(zipPath string, fn func(AggTrade) error) error {
	return IterZipCSV(zipPath, func(rec []string) error {
		tr, err := ParseAggRow(rec)
		if err != nil {
			return nil
		}
		return fn(tr)
	})
}

func iterZipCSV(zipPath string, skipHeader bool, fn func(row []string) error) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		if !strings.HasSuffix(strings.ToLower(f.Name), ".csv") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		cr := csv.NewReader(bufio.NewReader(rc))
		cr.FieldsPerRecord = -1
		for {
			rec, err := cr.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				_ = rc.Close()
				return err
			}
			if skipHeader && IsHeaderRow(rec) {
				continue
			}
			if err := fn(rec); err != nil {
				_ = rc.Close()
				return err
			}
		}
		_ = rc.Close()
	}
	return nil
}

func IsHeaderRow(rec []string) bool {
	if len(rec) == 0 {
		return true
	}
	s := strings.ToLower(strings.TrimSpace(rec[0]))
	if strings.Contains(s, "agg") && strings.Contains(s, "trade") {
		return true
	}
	if _, err := strconv.ParseInt(s, 10, 64); err != nil {
		return true
	}
	return false
}

func ParseAggRow(rec []string) (AggTrade, error) {
	if len(rec) < 7 {
		return AggTrade{}, fmt.Errorf("short row")
	}
	id, _ := strconv.ParseInt(rec[0], 10, 64)
	px, err := strconv.ParseFloat(rec[1], 64)
	if err != nil || px <= 0 {
		return AggTrade{}, fmt.Errorf("invalid price")
	}
	qty, err := strconv.ParseFloat(rec[2], 64)
	if err != nil || qty <= 0 {
		return AggTrade{}, fmt.Errorf("invalid qty")
	}
	first, _ := strconv.ParseInt(rec[3], 10, 64)
	last, _ := strconv.ParseInt(rec[4], 10, 64)
	ms, err := strconv.ParseInt(rec[5], 10, 64)
	if err != nil {
		return AggTrade{}, err
	}
	maker := strings.EqualFold(rec[6], "true") || rec[6] == "1"
	return AggTrade{
		ID: id, Price: px, Qty: qty, FirstTradeID: first, LastTradeID: last,
		Time: time.UnixMilli(ms).UTC(), BuyerMaker: maker,
	}, nil
}

func (tr AggTrade) Event(symbol string) md.Event {
	return md.Event{
		Kind: md.KindTrade, EventTime: tr.Time, ReceiveTime: tr.Time,
		Provider: "binance_vision", Venue: "binance_usdm", Instrument: symbol, Seq: tr.ID,
		Payload: md.Trade{Price: tr.Price, Qty: tr.Qty, BuyerMaker: tr.BuyerMaker},
		Prov: md.Provenance{
			Provider: "binance_vision", Venue: "binance_usdm", Instrument: symbol,
			Quality: md.QualityDirect, FreshnessClass: md.FreshDirect,
			EventTime: tr.Time, ReceivedAt: tr.Time, AvailableAt: tr.Time,
		},
	}
}

func Aggressor(buyerMaker bool) string { return flow.ClassifyAggressor(buyerMaker) }

type Quality struct {
	Rows, Duplicates, OutOfOrder, Invalid, Gaps, Headers int
}

func (a *Archive) ScanDay(zipPath string) (Quality, error) {
	var q Quality
	var lastID int64
	var lastT time.Time
	err := iterZipCSV(zipPath, false, func(rec []string) error {
		if IsHeaderRow(rec) {
			q.Headers++
			return nil
		}
		tr, err := ParseAggRow(rec)
		if err != nil {
			q.Invalid++
			return nil
		}
		q.Rows++
		if lastID != 0 && tr.ID == lastID {
			q.Duplicates++
		}
		if lastID != 0 && tr.ID < lastID {
			q.OutOfOrder++
		}
		if lastID != 0 && tr.ID > lastID+1 {
			q.Gaps++
		}
		if !lastT.IsZero() && tr.Time.Before(lastT) {
			q.OutOfOrder++
		}
		lastID, lastT = tr.ID, tr.Time
		return nil
	})
	return q, err
}

func DaysInclusive(from, to time.Time) []string {
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	to = time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)
	var out []string
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		out = append(out, d.Format("2006-01-02"))
	}
	return out
}
