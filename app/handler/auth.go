package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"screenshot-api/model"
	"screenshot-api/storage"
)

type AuthHandler struct {
	storage *storage.Storage
}

func NewAuthHandler(s *storage.Storage) *AuthHandler {
	return &AuthHandler{storage: s}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	User   *model.User   `json:"user"`
	APIKey *model.APIKey `json:"api_key"`
}

type loginResponse struct {
	APIKey *model.APIKey `json:"api_key"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		jsonError(w, "email and password required", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	user, err := h.storage.CreateUser(req.Email, string(hash))
	if err != nil {
		log.Printf("CreateUser error: %v", err)
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	apiKey, err := h.storage.CreateAPIKey(user.ID)
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, registerResponse{User: user, APIKey: apiKey}, http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.storage.GetUserByEmail(req.Email)
	if err != nil {
		jsonError(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		jsonError(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	apiKey, err := h.storage.GetAPIKeyByUserID(user.ID)
	if err != nil {
		log.Printf("GetAPIKeyByUserID error: %v", err)
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, loginResponse{APIKey: apiKey}, http.StatusOK)
}

func jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func jsonResponse(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
