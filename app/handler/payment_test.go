package handler

import (
	"testing"

	"screenshot-api/model"
)

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

func TestResolveInvoiceAmountsSupportsBTCAndUSD(t *testing.T) {
	sats, currency, err := resolveInvoiceAmounts(invoiceRequest{AmountBTC: 0.25, Currency: "BTC"}, 10000, &model.CurrencyRate{RateToSatoshi: 100000000, RateToUSD: 1})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if currency != "BTC" {
		t.Fatalf("expected currency BTC, got %s", currency)
	}
	if sats != 25000000 {
		t.Fatalf("expected satoshi amount 25000000, got %d", sats)
	}
}
