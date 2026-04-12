package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"screenshot-api/config"
	"screenshot-api/handler"
	"screenshot-api/middleware"
	"screenshot-api/storage"
)

var (
	browserlessURL   string
	browserlessToken string
	globalStore      *storage.Storage
)

func main() {
	cfg := config.Load()
	browserlessURL = cfg.BrowserlessURL
	browserlessToken = cfg.BrowserlessToken

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

	http.HandleFunc("/auth/register", authHandler.Register)
	http.HandleFunc("/auth/login", authHandler.Login)

	http.HandleFunc("/payments/create", authMiddleware.Authenticate(paymentHandler.CreateInvoice))
	http.HandleFunc("/internal/confirm-payment", confirmPaymentHandler)
	http.HandleFunc("/screenshot", authMiddleware.Authenticate(screenshotHandler))

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

	userID, amountUSD, err := globalStore.ConfirmInvoice(address)
	if err != nil {
		http.Error(w, "invoice not found or already confirmed", http.StatusNotFound)
		return
	}

	if err := globalStore.UpdateUserBalance(userID, amountUSD); err != nil {
		http.Error(w, "failed to update balance", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("balance updated"))
}

func init() {
	_ = os.Getenv
}
