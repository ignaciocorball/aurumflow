package research

import "strconv"

func strconvFormat(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}
