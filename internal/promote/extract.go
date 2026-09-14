package promote

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"aurumflow/internal/chronosplit"
	"aurumflow/internal/costmodel"
	"aurumflow/internal/legacynorm"
	"aurumflow/internal/mktval"
	"aurumflow/pkg/models"
)

func HashCandles(cs []models.Candle) string {
	h := sha256.New()
	for _, c := range cs {
		fmt.Fprintf(h, "%s %.8f %.8f %.8f %.8f\n", c.Time.UTC().Format(time.RFC3339Nano), c.Open, c.High, c.Low, c.Close)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func Extract(market string, m15, h1, h4 []models.Candle, spread float64) Evidence {
	ev := Evidence{
		Market:         market,
		BrokerIdentity: "capital.com",
		Spread:         spread,
		DataQuality:    "INVALID",
		LiveShadow:     "SHADOW",
		DemoStatus:     "NO",
	}
	if len(m15) == 0 {
		ev.ResearchStatus = mktval.ResearchRejected
		return ev
	}
	ev.DatasetHash = HashCandles(m15)
	if len(m15) >= 200 {
		ev.DataQuality = "VALID"
	} else {
		ev.DataQuality = "INSUFFICIENT"
	}
	if len(m15) < 80 {
		ev.ResearchStatus = mktval.DataReady
		return ev
	}
	split := chronosplit.Of(m15[0].Time, m15[len(m15)-1].Time)
	norm := legacynorm.ScanNormalized(m15, h1, h4)
	ev.Signals = len(norm)
	normal := mktval.Holdout(mktval.Simulate(norm, m15, spread, costmodel.Normal, split))
	stress := mktval.Holdout(mktval.Simulate(norm, m15, spread, costmodel.Stress, split))
	ns := mktval.Summarize(normal)
	ss := mktval.Summarize(stress)
	ev.HoldoutN = ns.N
	ev.NormalExp = ns.Expectancy
	ev.StressExp = ss.Expectancy
	ev.Hit = ns.Hit
	ev.ProfitFactor = ns.PF
	ev.MaxDD = ns.MaxDD
	ev.MFE = ns.MFE
	ev.MAE = ns.MAE
	ev.LongExp = ns.LongExp
	ev.ShortExp = ns.ShortExp
	ev.PositiveThirds = ns.PositiveThirds
	ev.LargestShare = ns.LargestShare
	ev.OutlierDominated = ns.OutlierDom
	ev.HoldoutStart = ns.First
	ev.HoldoutEnd = ns.Last
	ev.ResearchStatus = mktval.DecideStatus(
		mktval.Row{Market: market, M15: len(m15), Signals: ev.Signals, HoldoutN: ev.HoldoutN},
		ns, ev.DataQuality == "VALID" || ev.DataQuality == "INSUFFICIENT", true,
	)
	return ev
}

func HoldoutDates(ev Evidence) string {
	if ev.HoldoutStart.IsZero() || ev.HoldoutEnd.IsZero() {
		return "n/a"
	}
	return ev.HoldoutStart.UTC().Format("2006-01-02") + " → " + ev.HoldoutEnd.UTC().Format("2006-01-02")
}
