package md

const (
	LossLossless = "LOSSLESS_REQUIRED"
	LossTolerant = "LOSS_TOLERANT"
	LossDisplay  = "DISPLAY_ONLY"

	ConsumerBus      = "bus"
	ConsumerBook     = "book"
	ConsumerMicro    = "microflow"
	ConsumerRadar    = "radar"
	ConsumerRecorder = "recorder"
	ConsumerRawbuf   = "rawbuf"
	ConsumerUI       = "ui"

	ReasonBackpressure = "backpressure"
	ReasonCtxCancel    = "ctx_cancel"
)

// ClassifyLoss is a bounded enum: consumer × kind → class.
func ClassifyLoss(consumer string, kind Kind) string {
	switch consumer {
	case ConsumerUI:
		return LossDisplay
	case ConsumerBook:
		if kind == KindBookDelta || kind == KindBookSnapshot {
			return LossLossless
		}
		return LossTolerant
	case ConsumerRecorder, ConsumerRawbuf, ConsumerBus:
		return LossLossless
	case ConsumerMicro, ConsumerRadar:
		if kind == KindTrade {
			return LossLossless
		}
		return LossTolerant
	default:
		return LossLossless
	}
}
