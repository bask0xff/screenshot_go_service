package handler

import (
	"testing"

	"screenshot-api/model"
)

func TestResolveInvoiceAmountsSupportsCustomCurrencyRate(t *testing.T) {
	rate := &model.CurrencyRate{CurrencyCode: "EUR", RateToUSD: 1.1, RateToSatoshi: 55000000}
	sats, currency, err := resolveInvoiceAmounts(invoiceRequest{Amount: 100, Currency: "EUR"}, 10000, rate)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if currency != "EUR" {
		t.Fatalf("expected currency EUR, got %s", currency)
	}
	if sats != 100 {
		t.Fatalf("expected satoshi amount 100, got %d", sats)
	}
}

func TestResolveInvoiceAmountsFallsBackToUSDWhenCurrencyMissing(t *testing.T) {
	sats, currency, err := resolveInvoiceAmounts(invoiceRequest{AmountUSD: 50}, 10000, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if currency != "USD" {
		t.Fatalf("expected default currency USD, got %s", currency)
	}
	if sats != 500000 {
		t.Fatalf("expected satoshi amount 500000, got %d", sats)
	}
}
