package optionsctx

const (
	Realtime   = "NOT_CONNECTED"
	Historical = "NOT_CONNECTED"
)

type Provider struct {
	Realtime   string
	Historical string
	Quality    string
}

func New() *Provider {
	return &Provider{Realtime: Realtime, Historical: Historical, Quality: "NOT_CONNECTED"}
}
