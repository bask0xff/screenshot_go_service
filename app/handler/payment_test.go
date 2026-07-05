package handler

import "testing"

func TestCalculateDiscountedAmount(t *testing.T) {
	got := calculateDiscountedAmount(100, 10)
	if got != 90 {
		t.Fatalf("expected 90, got %.2f", got)
	}
}

func TestNormalizePaymentMethod(t *testing.T) {
	if got := normalizePaymentMethod("BTC"); got != "bitcoin" {
		t.Fatalf("expected bitcoin, got %s", got)
	}

	if got := normalizePaymentMethod("credit-card"); got != "card" {
		t.Fatalf("expected card, got %s", got)
	}
}
