package main

import (
	"context"
	"testing"
)

func TestMarketAllowsExecution(t *testing.T) {
	if !marketAllowsExecution("TRADEABLE") {
		t.Fatal("TRADEABLE")
	}
	if marketAllowsExecution("CLOSED") || marketAllowsExecution("") || marketAllowsExecution("OFFLINE") {
		t.Fatal("closed/empty must block")
	}
}

func TestDiscoverTradeableEpic_Explicit(t *testing.T) {
	epic, how, err := discoverTradeableEpic(context.Background(), nil, "GOLD")
	if err != nil || epic != "GOLD" || how != "explicit" {
		t.Fatalf("%s %s %v", epic, how, err)
	}
}

func TestCanaryDealSize(t *testing.T) {
	sz, err := canaryDealSize(0.01, 0.01, 50000)
	if err != nil || sz != 0.01 {
		t.Fatalf("got %v %v", sz, err)
	}
	if _, err := canaryDealSize(0, 0.01, 1); err == nil {
		t.Fatal("zero min")
	}
	if _, err := canaryDealSize(0.01, 0, 1); err == nil {
		t.Fatal("zero increment")
	}
}
