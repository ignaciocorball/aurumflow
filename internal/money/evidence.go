package money

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"time"
)

type journalLine struct {
	AccountCurrency string    `json:"account_currency"`
	Ask             float64   `json:"ask"`
	Bid             float64   `json:"bid"`
	CloseablePrice  float64   `json:"closeablePrice"`
	DealID          string    `json:"dealId"`
	Direction       string    `json:"direction"`
	Size            float64   `json:"size"`
	UPL             float64   `json:"upl"`
	Instrument      string    `json:"instrument"`
	EventTime       time.Time `json:"event_time"`
	ReceiveTime     time.Time `json:"receive_time"`
}

func EvidenceComplete(s MonetaryInstrumentSpec) bool {
	return s.ValidationStatus == RuntimeValidated &&
		s.MoneyPerPriceUnit > 0 &&
		strings.TrimSpace(s.MoneyPerPriceUnitCurrency) != "" &&
		s.CalibrationSamples >= MinCalibSamples &&
		s.PriceRange > 0 &&
		s.UPLRange > 0 &&
		s.Estimator == EstimatorTheilSen &&
		strings.TrimSpace(s.MetadataStatus) != "" &&
		s.PositionClosed &&
		s.BrokerPositionsAfter == 0
}

func CompleteEvidence(spec MonetaryInstrumentSpec, res CalibrationResult, positionsAfter int) MonetaryInstrumentSpec {
	spec = ApplyRuntime(spec, res)
	spec.PositionClosed = positionsAfter == 0
	spec.BrokerPositionsAfter = positionsAfter
	return spec
}

func LoadCalibrationJournal(path string) ([]CalibrationSample, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []CalibrationSample
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var row journalLine
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			continue
		}
		out = append(out, CalibrationSample{
			Timestamp:      row.ReceiveTime,
			DealID:         row.DealID,
			Direction:      row.Direction,
			Size:           row.Size,
			Bid:            row.Bid,
			Ask:            row.Ask,
			CloseablePrice: row.CloseablePrice,
			BrokerUPL:      row.UPL,
			AccountCurrency: row.AccountCurrency,
			EventTime:      row.EventTime,
			ReceiveTime:    row.ReceiveTime,
		})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
