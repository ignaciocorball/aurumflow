package cmectx

const (
	NQ_MBO = "NOT_CONNECTED"
	GC_MBO = "NOT_CONNECTED"
	NQPub  = "NOT_CONNECTED"
	GCPub  = "NOT_CONNECTED"
)

type PublicProvider struct {
	Status string
}

func New() *PublicProvider {
	return &PublicProvider{Status: "NOT_CONNECTED"}
}

func Statuses() map[string]string {
	return map[string]string{
		"CME_NQ_MBO": NQ_MBO,
		"CME_GC_MBO": GC_MBO,
		"CME_NQ_PUBLIC": NQPub,
		"CME_GC_PUBLIC": GCPub,
		"OPTIONS_REALTIME": "NOT_CONNECTED",
		"OPTIONS_HISTORICAL": "NOT_CONNECTED",
	}
}
