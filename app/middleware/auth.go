package middleware

import (
	"net/http"

	"screenshot-api/storage"
)

type AuthMiddleware struct {
	storage *storage.Storage
}

func NewAuthMiddleware(s *storage.Storage) *AuthMiddleware {
	return &AuthMiddleware{storage: s}
}

func (m *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			http.Error(w, `{"error": "missing api key"}`, http.StatusUnauthorized)
			return
		}

		key, err := m.storage.GetAPIKey(apiKey)
		if err != nil {
			http.Error(w, `{"error": "invalid api key"}`, http.StatusUnauthorized)
			return
		}

		// Увеличить счётчик запросов
		m.storage.IncrementRequests(key.Key)

		next(w, r)
	}
}
