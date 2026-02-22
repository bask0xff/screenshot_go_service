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

var browserlessURL string

func main() {
	cfg := config.Load()

	browserlessURL = cfg.BrowserlessURL

	// Подключение к БД
	store, err := storage.New(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	log.Println("connected to database")

	// Запуск миграций
	if err := store.RunMigrations(cfg); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	log.Println("migrations applied")

	// Handlers и middleware
	authHandler := handler.NewAuthHandler(store)
	authMiddleware := middleware.NewAuthMiddleware(store)

	// Роуты
	http.HandleFunc("/auth/register", authHandler.Register)
	http.HandleFunc("/auth/login", authHandler.Login)
	http.HandleFunc("/screenshot", authMiddleware.Authenticate(screenshotHandler))

	log.Println("server started on :8082")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func screenshotHandler(w http.ResponseWriter, r *http.Request) {
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

func init() {
	_ = os.Getenv // подавить предупреждение компилятора
}
