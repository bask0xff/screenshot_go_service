package handler

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"screenshot-api/model"
	"screenshot-api/storage"
	"strings"
	"time"
)

type PaymentHandler struct {
	storage *storage.Storage
}

func NewPaymentHandler(s *storage.Storage) *PaymentHandler {
	return &PaymentHandler{storage: s}
}

type invoiceRequest struct {
	Amount        int64  `json:"amount,omitempty"`
	PaymentMethod string `json:"payment_method,omitempty"`
	Currency      string `json:"currency,omitempty"`
	PromoCode     string `json:"promo_code,omitempty"`
	IsTest        bool   `json:"is_test,omitempty"`
}

type invoiceResponse struct {
	ID             int     `json:"id"`
	Address        string  `json:"address"`
	AmountBTC      float64 `json:"amount_btc"`
	AmountUSD      float64 `json:"amount_usd"`
	Amount         int64   `json:"amount"`
	AmountCurrency string  `json:"amount_currency"`
	AmountPayable  int64   `json:"amount_payable"`
	PaymentMethod  string  `json:"payment_method"`
	Currency       string  `json:"currency"`
	PromoCode      string  `json:"promo_code,omitempty"`
	Status         string  `json:"status"`
	IsTest         bool    `json:"is_test"`
	ExpiresAt      string  `json:"expires_at"`
}

type cancelRequest struct {
	Address string `json:"address"`
}

type promoCodeRequest struct {
	Code            string  `json:"code"`
	DiscountPercent float64 `json:"discount_percent"`
	MaxUses         int     `json:"max_uses"`
	ExpiresAt       string  `json:"expires_at"`
}

func (h *PaymentHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	keyStr := r.Header.Get("X-API-Key")
	apiKey, err := h.storage.GetAPIKey(keyStr)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req invoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	paymentMethod := normalizePaymentMethod(req.PaymentMethod)
	if paymentMethod == "" {
		paymentMethod = "bitcoin"
	}

	btcPrice, err := getBTCPrice()
	if err != nil {
		jsonError(w, "failed to fetch exchange rate", http.StatusInternalServerError)
		return
	}

	rate, err := h.storage.GetCurrencyRate(currencyCodeOrDefault(req.Currency), btcPrice)
	if err != nil {
		jsonError(w, "failed to resolve currency rate", http.StatusInternalServerError)
		return
	}

	amountSatoshi, currency, err := resolveInvoiceAmounts(req, btcPrice, rate)
	if err != nil {
		jsonError(w, "invalid amount", http.StatusBadRequest)
		return
	}

	payableSatoshi := amountSatoshi
	promoCode := strings.ToUpper(strings.TrimSpace(req.PromoCode))
	if promoCode != "" {
		promo, err := h.storage.GetPromoCode(promoCode)
		if err == nil && promo.Active && promo.UsedCount < promo.MaxUses && time.Now().Before(promo.ExpiresAt) {
			payableSatoshi = int64(float64(amountSatoshi) * (1 - promo.DiscountPercent/100))
			if err := h.storage.UsePromoCode(promoCode); err != nil {
				log.Printf("failed to increment promo usage: %v", err)
			}
		} else {
			jsonError(w, "invalid or expired promo code", http.StatusBadRequest)
			return
		}
	}

	amountSatoshi = payableSatoshi

	addr, err := h.storage.GetRandomFreeAddress()
	if err != nil {
		jsonError(w, "no addresses available", http.StatusServiceUnavailable)
		return
	}

	invoice, err := h.storage.CreateInvoiceWithDetails(apiKey.UserID, addr, amountSatoshi, paymentMethod, currency, promoCode, "", req.IsTest)
	if err != nil {
		log.Printf("CRITICAL DATABASE ERROR in CreateInvoice: %v", err)
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}

	_ = h.storage.AddPaymentEvent(invoice.ID, "created", fmt.Sprintf("payment_method=%s", paymentMethod))

	amountUSD := satoshisToUSD(amountSatoshi, btcPrice)
	amountBTC := satoshisToBTC(amountSatoshi)

	jsonResponse(w, invoiceResponse{
		ID:             invoice.ID,
		Address:        invoice.Address,
		AmountBTC:      amountBTC,
		AmountUSD:      amountUSD,
		Amount:         amountSatoshi,
		AmountCurrency: currency,
		AmountPayable:  amountSatoshi,
		PaymentMethod:  paymentMethod,
		Currency:       currency,
		PromoCode:      promoCode,
		Status:         invoice.Status,
		IsTest:         req.IsTest,
		ExpiresAt:      invoice.ExpiresAt.Format(time.RFC3339),
	}, http.StatusCreated)
}

func (h *PaymentHandler) CreateTestInvoice(w http.ResponseWriter, r *http.Request) {
	keyStr := r.Header.Get("X-API-Key")
	apiKey, err := h.storage.GetAPIKey(keyStr)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req invoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.IsTest = true

	paymentMethod := normalizePaymentMethod(req.PaymentMethod)
	if paymentMethod == "" {
		paymentMethod = "bitcoin"
	}

	btcPrice, err := getBTCPrice()
	if err != nil {
		jsonError(w, "failed to fetch exchange rate", http.StatusInternalServerError)
		return
	}

	rate, err := h.storage.GetCurrencyRate(currencyCodeOrDefault(req.Currency), btcPrice)
	if err != nil {
		jsonError(w, "failed to resolve currency rate", http.StatusInternalServerError)
		return
	}

	amountSatoshi, currency, err := resolveInvoiceAmounts(req, btcPrice, rate)
	if err != nil {
		jsonError(w, "invalid amount", http.StatusBadRequest)
		return
	}
	addr, err := h.storage.GetRandomFreeAddress()
	if err != nil {
		jsonError(w, "no addresses available", http.StatusServiceUnavailable)
		return
	}

	invoice, err := h.storage.CreateInvoiceWithDetails(apiKey.UserID, addr, amountSatoshi, paymentMethod, currency, "", "", req.IsTest)
	if err != nil {
		log.Printf("CRITICAL DATABASE ERROR in CreateTestInvoice: %v", err)
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}

	_ = h.storage.AddPaymentEvent(invoice.ID, "test_created", fmt.Sprintf("payment_method=%s", paymentMethod))

	amountUSD := satoshisToUSD(amountSatoshi, btcPrice)
	amountBTC := satoshisToBTC(amountSatoshi)

	jsonResponse(w, invoiceResponse{
		ID:             invoice.ID,
		Address:        invoice.Address,
		AmountBTC:      amountBTC,
		AmountUSD:      amountUSD,
		Amount:         amountSatoshi,
		AmountCurrency: currency,
		AmountPayable:  amountSatoshi,
		PaymentMethod:  paymentMethod,
		Currency:       currency,
		Status:         invoice.Status,
		IsTest:         true,
		ExpiresAt:      invoice.ExpiresAt.Format(time.RFC3339),
	}, http.StatusCreated)
}

func (h *PaymentHandler) CancelInvoice(w http.ResponseWriter, r *http.Request) {
	var req cancelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Address == "" {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.storage.CancelInvoice(req.Address); err != nil {
		jsonError(w, "failed to cancel invoice", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]string{"status": "cancelled", "address": req.Address}, http.StatusOK)
}

func (h *PaymentHandler) CreatePromoCode(w http.ResponseWriter, r *http.Request) {
	var req promoCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DiscountPercent < 0 || req.DiscountPercent > 100 {
		jsonError(w, "invalid promo code payload", http.StatusBadRequest)
		return
	}

	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if code == "" {
		code = generatePromoCode(12)
	}
	if len(code) != 12 || !strings.ContainsAny(code, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789") {
		jsonError(w, "promo code must be 12 uppercase alphanumeric characters", http.StatusBadRequest)
		return
	}

	expiresAt, err := time.Parse(time.RFC3339, req.ExpiresAt)
	if err != nil {
		expiresAt = time.Now().Add(24 * time.Hour)
	}

	promo, err := h.storage.CreatePromoCode(code, req.DiscountPercent, req.MaxUses, expiresAt)
	if err != nil {
		jsonError(w, "failed to create promo code", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, promo, http.StatusCreated)
}

func generatePromoCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return strings.ToUpper(strings.TrimSpace(fmt.Sprintf("PROMO%02d", i+1)))
		}
		b[i] = charset[idx.Int64()]
	}
	return string(b)
}

func resolveInvoiceAmounts(req invoiceRequest, btcPrice float64, rate *model.CurrencyRate) (amountSatoshi int64, currency string, err error) {
	currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "USD"
	}

	if req.Amount > 0 {
		amountSatoshi = req.Amount
		return
	}

	return 0, currency, fmt.Errorf("invalid amount")
}

func satoshisToUSD(sats int64, btcPrice float64) float64 {
	return float64(sats) / 100000000 * btcPrice
}

func satoshisToBTC(sats int64) float64 {
	return float64(sats) / 100000000
}

func currencyCodeOrDefault(code string) string {
	trimmed := strings.ToUpper(strings.TrimSpace(code))
	if trimmed == "" {
		return "USD"
	}
	return trimmed
}

func calculateDiscountedAmount(amount float64, percent float64) float64 {
	return amount * (1 - percent/100)
}

func normalizePaymentMethod(method string) string {
	switch strings.ToLower(strings.TrimSpace(method)) {
	case "btc", "bitcoin", "crypto":
		return "bitcoin"
	case "card", "credit-card", "credit_card", "stripe":
		return "card"
	case "bank", "bank-transfer", "wire":
		return "bank"
	case "", "default":
		return "bitcoin"
	default:
		return strings.ToLower(strings.TrimSpace(method))
	}
}

func getBTCPrice() (float64, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var data map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	price, ok := data["bitcoin"]["usd"]
	if !ok {
		return 0, fmt.Errorf("price not found")
	}
	return price, nil
}
