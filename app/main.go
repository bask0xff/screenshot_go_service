package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"screenshot-api/config"
	"screenshot-api/handler"
	"screenshot-api/middleware"
	"screenshot-api/storage"
)

var (
	browserlessURL   string
	browserlessToken string
	bitcoinRPCUser   string
	bitcoinRPCPass   string
	bitcoinRPCHost   string
	bitcoinRPCPort   string
	globalStore      *storage.Storage
)

func main() {
	cfg := config.Load()
	browserlessURL = cfg.BrowserlessURL
	browserlessToken = cfg.BrowserlessToken
	bitcoinRPCUser = cfg.BitcoinRPCUser
	bitcoinRPCPass = cfg.BitcoinRPCPass
	bitcoinRPCHost = cfg.BitcoinRPCHost
	bitcoinRPCPort = cfg.BitcoinRPCPort

	store, err := storage.New(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	globalStore = store
	log.Println("connected to database")

	if err := store.RunMigrations(cfg); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	log.Println("migrations applied")

	authHandler := handler.NewAuthHandler(store)
	paymentHandler := handler.NewPaymentHandler(store)
	authMiddleware := middleware.NewAuthMiddleware(store)

	// CORS снаружи, Auth внутри — оба работают
	http.HandleFunc("/payments/create", corsMiddleware(authMiddleware.Authenticate(paymentHandler.CreateInvoice)))
	http.HandleFunc("/payments/test-create", corsMiddleware(authMiddleware.Authenticate(paymentHandler.CreateTestInvoice)))
	http.HandleFunc("/payments/cancel", corsMiddleware(authMiddleware.Authenticate(paymentHandler.CancelInvoice)))
	http.HandleFunc("/payments/promos/create", corsMiddleware(authMiddleware.Authenticate(paymentHandler.CreatePromoCode)))
	http.HandleFunc("/screenshot", corsMiddleware(authMiddleware.Authenticate(screenshotHandler)))

	// Эти роуты без Auth — так и должно быть, иначе не войдёшь
	http.HandleFunc("/auth/register", corsMiddleware(authHandler.Register))
	http.HandleFunc("/auth/login", corsMiddleware(authHandler.Login))
	http.HandleFunc("/internal/confirm-payment", corsMiddleware(confirmPaymentHandler))

	go func() {
		for {
			syncPendingBitcoinInvoices()
			time.Sleep(30 * time.Second)
		}
	}()

	log.Println("server started on :8082")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func screenshotHandler(w http.ResponseWriter, r *http.Request) {
	apiKeyStr := r.Header.Get("X-API-Key")
	key, err := globalStore.GetAPIKey(apiKeyStr)
	if err != nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	if key.Tier != "free" {
		balance, err := globalStore.GetUserBalance(key.UserID)
		if err != nil || balance < 0.10 {
			http.Error(w, `{"error": "insufficient balance. Please top up your account."}`, http.StatusPaymentRequired)
			return
		}
		_ = globalStore.UpdateUserBalance(key.UserID, -0.10)
	}

	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, `{"error": "url required"}`, http.StatusBadRequest)
		return
	}

	// Формат запроса для ghcr.io/browserless/chromium
	payload := map[string]interface{}{
		"url": targetURL,
		"options": map[string]interface{}{
			"fullPage": true,
			"type":     "png",
		},
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, browserlessURL, bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, `{"error": "failed to create request"}`, http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if browserlessToken != "" {
		req.Header.Set("Authorization", "Bearer "+browserlessToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("browserless error: %v", err)
		http.Error(w, `{"error": "screenshot failed"}`, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("browserless returned %d: %s", resp.StatusCode, string(respBody))
		http.Error(w, `{"error": "screenshot failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	io.Copy(w, resp.Body)
}

func confirmPaymentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "address required", http.StatusBadRequest)
		return
	}

	invoice, err := globalStore.GetInvoiceByAddress(address)
	if err != nil || invoice == nil {
		http.Error(w, "invoice not found", http.StatusNotFound)
		return
	}

	confirmed, err := checkBitcoinPayment(address, satoshisToBTC(invoice.Amount))
	if err != nil {
		log.Printf("bitcoin confirmation check failed: %v", err)
		confirmed = false
	}

	if !confirmed {
		http.Error(w, "payment not confirmed yet", http.StatusAccepted)
		return
	}

	btcPrice, err := getBTCPrice()
	if err != nil {
		log.Printf("failed to fetch exchange rate: %v", err)
		http.Error(w, "failed to fetch exchange rate", http.StatusInternalServerError)
		return
	}

	userID, amountSatoshi, err := globalStore.ConfirmInvoice(address)
	if err != nil {
		http.Error(w, "invoice not found or already confirmed", http.StatusNotFound)
		return
	}

	if err := globalStore.UpdateUserBalance(userID, satoshisToUSD(amountSatoshi, btcPrice)); err != nil {
		http.Error(w, "failed to update balance", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("balance updated"))
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

func checkBitcoinPayment(address string, expectedBTC float64) (bool, error) {
	if bitcoinRPCUser == "" || bitcoinRPCPass == "" {
		return false, fmt.Errorf("bitcoin rpc is not configured")
	}

	payload := map[string]interface{}{
		"jsonrpc": "1.0",
		"id":      "curltext",
		"method":  "listtransactions",
		"params":  []interface{}{"*", 1000},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest(http.MethodPost, "http://"+bitcoinRPCHost+":"+bitcoinRPCPort, bytes.NewBuffer(body))
	if err != nil {
		return false, err
	}
	req.SetBasicAuth(bitcoinRPCUser, bitcoinRPCPass)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var rpcResp struct {
		Result []struct {
			Address       string  `json:"address"`
			Amount        float64 `json:"amount"`
			Category      string  `json:"category"`
			Confirmations int     `json:"confirmations"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return false, err
	}

	for _, tx := range rpcResp.Result {
		if tx.Address == address && tx.Category == "receive" && tx.Confirmations >= 1 && tx.Amount >= expectedBTC {
			return true, nil
		}
	}
	return false, nil
}

func syncPendingBitcoinInvoices() {
	if globalStore == nil {
		return
	}

	invoices, err := globalStore.ListPendingInvoices()
	if err != nil {
		log.Printf("failed to list pending invoices: %v", err)
		return
	}

	for _, invoice := range invoices {
		confirmed, err := checkBitcoinPayment(invoice.Address, satoshisToBTC(invoice.Amount))
		if err != nil {
			log.Printf("bitcoin sync failed for %s: %v", invoice.Address, err)
			continue
		}
		if !confirmed {
			continue
		}

		userID, amountSatoshi, err := globalStore.ConfirmInvoice(invoice.Address)
		if err != nil {
			log.Printf("failed to confirm invoice %s: %v", invoice.Address, err)
			continue
		}
		if err := globalStore.UpdateUserBalance(userID, satoshisToUSD(amountSatoshi, 1)); err != nil {
			log.Printf("failed to credit balance for invoice %s: %v", invoice.Address, err)
		}
	}
}

func init() {
	_ = os.Getenv
}
