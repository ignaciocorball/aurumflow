package exhaustion

import (
	"strings"
	"time"

	"aurumflow/internal/strategy"
)

type EvidenceItem struct {
	Name  string
	Value float64
	Note  string
}

type Snapshot struct {
	Timestamp           time.Time
	Instrument          string
	LegacyDirection     int
	LegacyScore         float64
	PressureScore       float64
	DirectionalPressure float64
	Classification      string
	Features            FeatureSet
	Confidence          float64
	Capabilities        uint32
	Evidence            []EvidenceItem
	BookCapability      string
	Mode                string
	ExhaustionEvidence  float64
	EvidenceStatus      string
}

type Engine struct {
	Instrument string
	Mode       string
	Caps       uint32
	trades     []Trade
	cvdHist    []float64
	magHist    []float64
	dispHist   []float64
	volHist    []float64
	intHist    []float64
	Last       Snapshot
}

func NewEngine(instrument string) *Engine {
	return &Engine{Instrument: instrument, Mode: ModeShadow, Caps: CapTradeFlow}
}

func (e *Engine) OnTrade(t time.Time, price, qty float64, buyerMaker bool) {
	e.trades = append(e.trades, Trade{T: t, Price: price, Qty: qty, BuyerMaker: buyerMaker})
	cut := t.Add(-20 * time.Minute)
	i := 0
	for i < len(e.trades) && e.trades[i].T.Before(cut) {
		i++
	}
	if i > 0 {
		e.trades = e.trades[i:]
	}
}

func (e *Engine) Observe(t0 time.Time, legacyDir int, legacyScore float64, pressure float64, atr float64) Snapshot {
	past := make([]Trade, 0, len(e.trades))
	for _, tr := range e.trades {
		if tr.T.Before(t0) {
			past = append(past, tr)
		}
	}
	cvd := 0.0
	for _, tr := range past {
		if tr.BuyerMaker {
			cvd -= tr.Qty
		} else {
			cvd += tr.Qty
		}
	}
	wins := []time.Duration{30 * time.Second, time.Minute, 3 * time.Minute, 5 * time.Minute, 15 * time.Minute}
	fs := FeatureSet{At: t0, Windows: map[string]FlowWindow{}, Prices: map[string]PriceWindow{}, CVD: cvd, LegacyDir: legacyDir, LegacyScore: int(legacyScore), ATR: atr}
	var flow5 FlowWindow
	var px5 PriceWindow
	for _, w := range wins {
		fw := FlowFrom(past, t0, w, 0)
		pw := PriceFrom(past, t0, w, legacyDir, fw.NetVol)
		fs.Windows[WindowKey(w)] = fw
		fs.Prices[WindowKey(w)] = pw
		if w == 5*time.Minute {
			flow5, px5 = fw, pw
		}
	}
	e.cvdHist = append(e.cvdHist, cvd)
	e.magHist = append(e.magHist, mathAbs(flow5.NetNotional))
	e.dispHist = append(e.dispHist, px5.DispVsFlow)
	e.volHist = append(e.volHist, px5.AbsReturn)
	e.intHist = append(e.intHist, flow5.TradeVelocity)
	trim(&e.cvdHist, 60)
	trim(&e.magHist, 60)
	trim(&e.dispHist, 60)
	trim(&e.volHist, 60)
	trim(&e.intHist, 60)
	// rolling stats exclude current observation (past-only baseline)
	magBase, dispBase := e.magHist, e.dispHist
	if len(magBase) > 1 {
		magBase = magBase[:len(magBase)-1]
		dispBase = dispBase[:len(dispBase)-1]
	}
	medM, madM := RollingMedianMAD(magBase)
	medD, madD := RollingMedianMAD(dispBase)
	fs.FlowMagNorm = RobustZ(mathAbs(flow5.NetNotional), medM, madM)
	fs.PriceDispNorm = RobustZ(px5.DispVsFlow, medD, madD)
	fs.FlowEffNorm = fs.PriceDispNorm - fs.FlowMagNorm
	fs.ImpactFailure = fs.FlowMagNorm - fs.PriceDispNorm
	cvdBase := e.cvdHist
	if len(cvdBase) > 1 {
		cvdBase = cvdBase[:len(cvdBase)-1]
	}
	fs.CVDPct = Percentile(cvdBase, cvd)
	fs.CVDZ = ZScore(cvdBase, cvd)
	fs.RealizedVol = px5.AbsReturn
	if len(e.volHist) > 1 {
		fs.ATRPct = Percentile(e.volHist[:len(e.volHist)-1], px5.AbsReturn)
	}
	if len(e.intHist) > 1 {
		fs.IntensityPct = Percentile(e.intHist[:len(e.intHist)-1], flow5.TradeVelocity)
	}
	fs.UTCHour = t0.UTC().Hour()
	fs.Dow = int(t0.UTC().Weekday())
	fs.Session = strings.Join(strategy.GetCurrentSession(t0), "+")
	if atr > 0 {
		fs.DistInvalidATR = 1
	}
	class := ClassifyV1(legacyDir, pressure)
	dp := DirectionalPressure(legacyDir, pressure)
	evScore := 0.0
	if class == ClassExhaustion {
		evScore = clamp01((fs.ImpactFailure + 2) / 4)
	}
	s := Snapshot{
		Timestamp: t0, Instrument: e.Instrument,
		LegacyDirection: legacyDir, LegacyScore: legacyScore,
		PressureScore: pressure, DirectionalPressure: dp,
		Classification: class, Features: fs,
		Confidence: 70, Capabilities: e.Caps, Mode: e.Mode,
		BookCapability: "BOOK_CAPABILITY_LIMITED",
		ExhaustionEvidence: evScore, EvidenceStatus: "EXPERIMENTAL_RESEARCH_ONLY",
		Evidence: []EvidenceItem{
			{Name: "DirectionalPressure", Value: dp, Note: "D*P unchanged PressureScore"},
			{Name: "FlowMagNorm", Value: fs.FlowMagNorm, Note: "past-only MAD"},
			{Name: "PriceDispNorm", Value: fs.PriceDispNorm, Note: "vs aggressive flow"},
			{Name: "ImpactFailure", Value: fs.ImpactFailure, Note: "high = high flow + low displacement"},
			{Name: "FlowEfficiency", Value: fs.FlowEffNorm, Note: "dispNorm - magNorm"},
		},
	}
	if e.Caps&CapBook == 0 {
		s.BookCapability = "BOOK_CAPABILITY_LIMITED"
	}
	e.Last = s
	return s
}

func trim(xs *[]float64, n int) {
	if len(*xs) > n {
		*xs = (*xs)[len(*xs)-n:]
	}
}

func mathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
