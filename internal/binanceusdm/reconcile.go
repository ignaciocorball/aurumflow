package binanceusdm

// FirstApplicable reports whether a buffered depth event may be applied after a REST snapshot.
// Binance USD-M: the first processed event must satisfy U <= lastUpdateId+1 <= u.
func FirstApplicable(lastUpdateID, firstID, finalID int64) bool {
	return firstID <= lastUpdateID+1 && lastUpdateID+1 <= finalID
}

// Obsolete reports events that arrived before the snapshot id (u <= lastUpdateId).
func Obsolete(lastUpdateID, finalID int64) bool {
	return finalID <= lastUpdateID
}
