package legacycompat

// Compatible != validated. Do not change Legacy strategy.
const (
	Compatible   = "LEGACY_COMPATIBLE"
	NeedsConfig  = "LEGACY_NEEDS_CONFIG"
	NotValidated = "LEGACY_NOT_VALIDATED"
	Unsupported  = "LEGACY_UNSUPPORTED"
)

type Entry struct {
	Market     string
	Status     string
	Validated  bool
	Blocker    string
	Notes      string
}

func Of(market string) Entry {
	switch market {
	case "GOLD":
		return Entry{Market: market, Status: Compatible, Validated: false, Notes: "ATR 8-18 / London+NY / GOLD-scale stops. Compatible != validated."}
	case "SILVER", "OIL_CRUDE":
		return Entry{Market: market, Status: NeedsConfig, Blocker: "ATR absolute buckets and pip/tick scale are GOLD-fitted", Notes: "needs instrument config; do not retune vs returns"}
	case "US100", "US500", "US30":
		return Entry{Market: market, Status: NeedsConfig, Blocker: "index point ATR != GOLD buckets; US cash session", Notes: "session + ATR normalization required"}
	case "DE40", "UK100":
		return Entry{Market: market, Status: NeedsConfig, Blocker: "Europe session + index ATR scale", Notes: "London session overlaps but ATR unvalidated"}
	case "J225", "CN50":
		return Entry{Market: market, Status: NeedsConfig, Blocker: "Asia hours vs London/NY hard filter", Notes: "CanTrade London/NY would skip Asia open"}
	case "BTC":
		return Entry{Market: market, Status: Unsupported, Blocker: "microstructure lab, not Legacy execution venue", Notes: "no Capital BTC execution in P8.2"}
	default:
		return Entry{Market: market, Status: NotValidated, Blocker: "no audit"}
	}
}

func AllCanonical() []Entry {
	ids := []string{"GOLD", "SILVER", "OIL_CRUDE", "US100", "US500", "US30", "DE40", "UK100", "J225", "CN50", "BTC"}
	out := make([]Entry, 0, len(ids))
	for _, id := range ids {
		out = append(out, Of(id))
	}
	return out
}
