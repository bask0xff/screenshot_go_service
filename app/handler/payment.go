package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"screenshot-api/storage"
	"time"
)

type PaymentHandler struct {
	storage *storage.Storage
}

func NewPaymentHandler(s *storage.Storage) *PaymentHandler {
	return &PaymentHandler{storage: s}
}

type invoiceRequest struct {
	AmountUSD float64 `json:"amount_usd"` // Сколько пользователь хочет зачислить
}

type invoiceResponse struct {
	Address   string  `json:"address"`
	AmountBTC float64 `json:"amount_btc"`
	AmountUSD float64 `json:"amount_usd"`
	ExpiresAt string  `json:"expires_at"`
}

func (h *PaymentHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	// 1. Извлекаем API-Key из заголовка (упрощенно, для получения UserID)
	keyStr := r.Header.Get("X-API-Key")
	apiKey, err := h.storage.GetAPIKey(keyStr)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req invoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.AmountUSD <= 0 {
		jsonError(w, "invalid amount", http.StatusBadRequest)
		return
	}

	// 2. Получаем курс BTC/USD (через CoinGecko)
	btcPrice, err := getBTCPrice()
	if err != nil {
		jsonError(w, "failed to fetch exchange rate", http.StatusInternalServerError)
		return
	}

	btcAmount := req.AmountUSD / btcPrice

	// 3. Берем свободный адрес из твоей таблицы btcaddress2
	addr, err := h.storage.GetRandomFreeAddress()
	if err != nil {
		jsonError(w, "no addresses available", http.StatusServiceUnavailable)
		return
	}

	// 4. Записываем инвойс в базу
	err = h.storage.CreateInvoice(apiKey.UserID, addr, req.AmountUSD, btcAmount)
	if err != nil {
		log.Printf("CRITICAL DATABASE ERROR in CreateInvoice: %v", err)

		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}

	// 5. Ответ пользователю
	jsonResponse(w, invoiceResponse{
		Address:   addr,
		AmountBTC: btcAmount,
		AmountUSD: req.AmountUSD,
		ExpiresAt: time.Now().Add(3 * time.Hour).Format(time.RFC3339),
	}, http.StatusCreated)
}

// Вспомогательная функция для получения курса
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
