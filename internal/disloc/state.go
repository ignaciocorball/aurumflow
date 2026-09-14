package disloc

const (
	Normal   = "NORMAL"
	Elevated = "ELEVATED"
	Severe   = "SEVERE"
	Extreme  = "EXTREME"
)

type Input struct {
	HighSalience int
	MedianSal    float64
	SpreadWide   int
}

func Classify(in Input) string {
	if in.HighSalience >= 6 && in.MedianSal >= 70 {
		return Extreme
	}
	if in.HighSalience >= 4 && in.MedianSal >= 60 {
		return Severe
	}
	if in.HighSalience >= 2 || in.MedianSal >= 55 {
		return Elevated
	}
	return Normal
}
