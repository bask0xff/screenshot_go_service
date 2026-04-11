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
	browserlessURL string
	globalStore    *storage.Storage // Сделаем store доступным для screenshotHandler
)

func main() {
	cfg := config.Load()
	browserlessURL = cfg.BrowserlessURL

	// Подключение к БД
	store, err := storage.New(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	globalStore = store
	log.Println("connected to database")

	// Запуск миграций
	if err := store.RunMigrations(cfg); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	log.Println("migrations applied")

	// Инициализация Handlers
	authHandler := handler.NewAuthHandler(store)
	paymentHandler := handler.NewPaymentHandler(store) // Добавили платежи
	authMiddleware := middleware.NewAuthMiddleware(store)

	// Роуты авторизации
	http.HandleFunc("/auth/register", authHandler.Register)
	http.HandleFunc("/auth/login", authHandler.Login)

	// Роуты платежей (защищены API-ключом)
	http.HandleFunc("/payments/create", authMiddleware.Authenticate(paymentHandler.CreateInvoice))

	// Внутренний роут для твоего ноутбука (защити его через секретный заголовок в проде)
	http.HandleFunc("/internal/confirm-payment", confirmPaymentHandler)

	// Основной функционал
	http.HandleFunc("/screenshot", authMiddleware.Authenticate(screenshotHandler))

	log.Println("server started on :8082")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func screenshotHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Извлекаем API-ключ, чтобы проверить баланс пользователя
	apiKeyStr := r.Header.Get("X-API-Key")
	key, err := globalStore.GetAPIKey(apiKeyStr)
	if err != nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// 2. Логика монетизации: если не Free Tier, проверяем баланс
	if key.Tier != "free" {
		balance, err := globalStore.GetUserBalance(key.UserID)
		if err != nil || balance < 0.10 { // Допустим, 1 скриншот стоит $0.10
			http.Error(w, `{"error": "insufficient balance. Please top up your account."}`, http.StatusPaymentRequired)
			return
		}
		// Списываем деньги за запрос
		_ = globalStore.UpdateUserBalance(key.UserID, -0.10)
	}

	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, `{"error": "url required"}`, http.StatusBadRequest)
		return
	}

	payload := map[string]interface{}{
		"url": targetURL,
		"options": map[string]interface{}{
			"fullPage": true,
			"type":     "png",
		},
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(browserlessURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, `{"error": "screenshot failed"}`, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "image/png")
	io.Copy(w, resp.Body)
}

// confirmPaymentHandler принимает сигнал от твоего ноутбука
func confirmPaymentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// ВАЖНО: Добавь проверку секретного токена из .env, чтобы никто не накрутил себе баланс
	// secret := r.Header.Get("X-Internal-Secret")

	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "address required", http.StatusBadRequest)
		return
	}

	// Подтверждаем инвойс и получаем userID и сумму
	userID, amountUSD, err := globalStore.ConfirmInvoice(address)
	if err != nil {
		http.Error(w, "invoice not found or already confirmed", http.StatusNotFound)
		return
	}

	// Начисляем баланс пользователю
	err = globalStore.UpdateUserBalance(userID, amountUSD)
	if err != nil {
		http.Error(w, "failed to update balance", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("balance updated"))
}

func init() {
	_ = os.Getenv
}
