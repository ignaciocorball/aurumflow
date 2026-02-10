package strategy

import "testing"

func TestH1Allowed(t *testing.T) {
	const th = 6
	tests := []struct {
		name             string
		blockH1Range     bool
		crypto           bool
		h1RangeExtraScore int
		trendH1          string
		direction        string
		finalScore       int
		scoreThreshold   int
		want             bool
	}{
		{
			name: "crypto_false_RANGE_BUY_high_score_rejected",
			blockH1Range: false, crypto: false, h1RangeExtraScore: 1,
			trendH1: "RANGE", direction: "BUY", finalScore: 9, scoreThreshold: th,
			want: false,
		},
		{
			name: "crypto_true_RANGE_BUY_finalScore_threshold_plus_1_allowed",
			blockH1Range: false, crypto: true, h1RangeExtraScore: 1,
			trendH1: "RANGE", direction: "BUY", finalScore: 7, scoreThreshold: th,
			want: true,
		},
		{
			name: "crypto_true_RANGE_BUY_finalScore_threshold_rejected",
			blockH1Range: false, crypto: true, h1RangeExtraScore: 1,
			trendH1: "RANGE", direction: "BUY", finalScore: 6, scoreThreshold: th,
			want: false,
		},
		{
			name: "crypto_true_BULLISH_BUY_allowed",
			blockH1Range: false, crypto: true, h1RangeExtraScore: 1,
			trendH1: "BULLISH", direction: "BUY", finalScore: 6, scoreThreshold: th,
			want: true,
		},
		{
			name: "crypto_true_BEARISH_SELL_allowed",
			blockH1Range: false, crypto: true, h1RangeExtraScore: 1,
			trendH1: "BEARISH", direction: "SELL", finalScore: 6, scoreThreshold: th,
			want: true,
		},
		{
			name: "blockH1Range_true_RANGE_rejected_even_with_crypto",
			blockH1Range: true, crypto: true, h1RangeExtraScore: 1,
			trendH1: "RANGE", direction: "BUY", finalScore: 8, scoreThreshold: th,
			want: false,
		},
		{
			name: "crypto_true_RANGE_extra_0_treated_as_1_minScore_7",
			blockH1Range: false, crypto: true, h1RangeExtraScore: 0,
			trendH1: "RANGE", direction: "BUY", finalScore: 7, scoreThreshold: th,
			want: true,
		},
		{
			name: "crypto_true_RANGE_extra_2_minScore_8",
			blockH1Range: false, crypto: true, h1RangeExtraScore: 2,
			trendH1: "RANGE", direction: "SELL", finalScore: 8, scoreThreshold: th,
			want: true,
		},
		{
			name: "crypto_true_RANGE_extra_2_finalScore_7_rejected",
			blockH1Range: false, crypto: true, h1RangeExtraScore: 2,
			trendH1: "RANGE", direction: "SELL", finalScore: 7, scoreThreshold: th,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := H1Allowed(tt.blockH1Range, tt.crypto, tt.h1RangeExtraScore, tt.trendH1, tt.direction, tt.finalScore, tt.scoreThreshold)
			if got != tt.want {
				t.Errorf("H1Allowed() = %v, want %v", got, tt.want)
			}
		})
	}
}
