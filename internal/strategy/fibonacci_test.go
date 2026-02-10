package strategy

import (
	"aurumflow/pkg/models"
	"testing"
)

func TestCalculateFibAlwaysLowLessThanHigh(t *testing.T) {
	// Last swing is LOW (index 10), previous is HIGH (index 5). Low price 105 > High price 100.
	// Before fix: fib.Low could be 105, fib.High 100 (inverted). After fix: fib.Low=100, fib.High=105.
	swings := []models.Swing{
		{Index: 5, Price: 100, Type: "HIGH"},
		{Index: 10, Price: 105, Type: "LOW"},
	}
	fib, ok := CalculateFib(swings)
	if !ok {
		t.Fatal("CalculateFib expected true")
	}
	if fib.Low >= fib.High {
		t.Errorf("Fib must have Low < High; got Low=%v High=%v", fib.Low, fib.High)
	}
	if fib.ImpulseDirection != "DOWN" {
		t.Errorf("Last swing was low (index 10 > 5), expected ImpulseDirection=DOWN; got %q", fib.ImpulseDirection)
	}
	// Levels should sit between Low and High
	if fib.Level382 < fib.Low || fib.Level382 > fib.High {
		t.Errorf("Level382 should be between Low and High; got %v", fib.Level382)
	}
}

func TestCalculateFibImpulseUP(t *testing.T) {
	// Last swing is HIGH (index 10), previous is LOW (index 5). fib.Low=100, fib.High=105.
	swings := []models.Swing{
		{Index: 5, Price: 100, Type: "LOW"},
		{Index: 10, Price: 105, Type: "HIGH"},
	}
	fib, ok := CalculateFib(swings)
	if !ok {
		t.Fatal("CalculateFib expected true")
	}
	if fib.Low >= fib.High {
		t.Errorf("Fib must have Low < High; got Low=%v High=%v", fib.Low, fib.High)
	}
	if fib.ImpulseDirection != "UP" {
		t.Errorf("Last swing was high (index 10 > 5), expected ImpulseDirection=UP; got %q", fib.ImpulseDirection)
	}
}
