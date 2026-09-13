package binancehist

import (
	"strconv"
	"time"

	"aurumflow/pkg/models"
)

func ParseKlineRow(rec []string) (models.Candle, error) {
	if len(rec) < 6 {
		return models.Candle{}, errShort
	}
	ms, err := strconv.ParseInt(rec[0], 10, 64)
	if err != nil {
		return models.Candle{}, err
	}
	o, _ := strconv.ParseFloat(rec[1], 64)
	h, _ := strconv.ParseFloat(rec[2], 64)
	l, _ := strconv.ParseFloat(rec[3], 64)
	c, _ := strconv.ParseFloat(rec[4], 64)
	v, _ := strconv.ParseFloat(rec[5], 64)
	if o <= 0 || h <= 0 || l <= 0 || c <= 0 {
		return models.Candle{}, errShort
	}
	return models.Candle{Time: time.UnixMilli(ms).UTC(), Open: o, High: h, Low: l, Close: c, Volume: v}, nil
}

var errShort = errInvalid("kline")

type errInvalid string

func (e errInvalid) Error() string { return string(e) + " invalid" }
