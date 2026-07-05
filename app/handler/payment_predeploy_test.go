package handler

import (
	"testing"

	"screenshot-api/model"
)

func TestResolveInvoiceAmountsSupportsCustomCurrencyRate(t *testing.T) {
	rate := &model.CurrencyRate{CurrencyCode: "EUR", RateToUSD: 1.1, RateToSatoshi: 55000000}
	usd, btc, sats, currency, err := resolveInvoiceAmounts(invoiceRequest{Amount: 100, Currency: "EUR"}, 10000, rate)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if currency != "EUR" {
		t.Fatalf("expected currency EUR, got %s", currency)
	}
	if usd != 110 {
		t.Fatalf("expected usd amount 110, got %.2f", usd)
	}
	if btc != 0.011 {
		t.Fatalf("expected btc amount 0.011, got %.8f", btc)
	}
	if sats != 5500000 {
		t.Fatalf("expected satoshi amount 5500000, got %d", sats)
	}
}

func TestResolveInvoiceAmountsFallsBackToUSDWhenCurrencyMissing(t *testing.T) {
	usd, btc, sats, currency, err := resolveInvoiceAmounts(invoiceRequest{AmountUSD: 50}, 10000, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if currency != "USD" {
		t.Fatalf("expected default currency USD, got %s", currency)
	}
	if usd != 50 {
		t.Fatalf("expected usd amount 50, got %.2f", usd)
	}
	if btc != 0.005 {
		t.Fatalf("expected btc amount 0.005, got %.8f", btc)
	}
	if sats != 500000 {
		t.Fatalf("expected satoshi amount 500000, got %d", sats)
	}
}
