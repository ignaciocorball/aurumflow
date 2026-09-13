package money

import (
	"testing"

	"aurumflow/internal/market"
)

func TestFromMarketDetails_LotSizeInference(t *testing.T) {
	s := FromMarketDetails(&market.MarketDetailsResponse{
		Instrument: market.MarketInstrument{Epic: "BTCUSD", Type: "CRYPTOCURRENCIES", Currency: "USD", LotSize: 1},
		DealingRules: market.DealingRules{
			MinDealSize: market.SizeRule{Value: 0.01}, MaxDealSize: market.SizeRule{Value: 10}, MinSizeIncrement: market.SizeRule{Value: 0.01},
		},
	})
	if !s.DealingComplete() || s.MoneyPerPriceUnit != 1 || s.ValidationStatus != BrokerMetadataOnly {
		t.Fatalf("%+v", s)
	}
	pnl, err := s.ExpectedPnL(0.01, 100, 110)
	if err != nil || pnl != 0.1 {
		t.Fatalf("pnl=%v err=%v", pnl, err)
	}
}

func TestExpectedPnL_Unknown(t *testing.T) {
	s := MonetaryInstrumentSpec{Epic: "GOLD"}
	if _, err := s.ExpectedPnL(1, 1, 2); err == nil {
		t.Fatal("unknown mpu")
	}
	if s.StrategyExecutable() {
		t.Fatal("unverified must not execute strategy")
	}
}

func TestPickRequiresDealingRules(t *testing.T) {
	s := FromMarketDetails(&market.MarketDetailsResponse{Instrument: market.MarketInstrument{Epic: "X"}})
	if s.DealingComplete() {
		t.Fatal("incomplete")
	}
}
