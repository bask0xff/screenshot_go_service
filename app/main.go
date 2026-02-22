package main

import (
    "bytes"
    "encoding/json"
    "io"
    "net/http"
    "os"
)

var apiKeys = map[string]string{
    "user_abc123": "free_tier",
    "user_xyz789": "paid_tier",
}

var browserlessURL = "http://browserless:3000/screenshot?token=" + os.Getenv("BROWSERLESS_TOKEN")

func screenshotHandler(w http.ResponseWriter, r *http.Request) {
    apiKey := r.Header.Get("X-API-Key")
    if _, ok := apiKeys[apiKey]; !ok {
        http.Error(w, `{"error": "invalid api key"}`, http.StatusUnauthorized)
        return
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

func main() {
    http.HandleFunc("/screenshot", screenshotHandler)
    http.ListenAndServe(":8082", nil)
}