package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"screenshot-api/model"
)

type memoryStore struct {
	users       map[string]*model.User
	apiKey      *model.APIKey
	methods     map[string]bool
	promos      map[string]*model.PromoCode
	lastInvoice *model.Invoice
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		users:   make(map[string]*model.User),
		apiKey:  &model.APIKey{ID: 1, UserID: 1, Key: "test-key", Tier: "free"},
		methods: map[string]bool{"bitcoin": true, "card": true, "bank": true},
		promos:  make(map[string]*model.PromoCode),
	}
}

func (s *memoryStore) CreateUser(email, hash string) (*model.User, error) {
	if _, exists := s.users[email]; exists {
		return nil, errors.New("duplicate email")
	}
	u := &model.User{ID: len(s.users) + 1, Email: email, Password: hash, CreatedAt: time.Now()}
	s.users[email] = u
	return u, nil
}
func (s *memoryStore) CreateAPIKey(userID int) (*model.APIKey, error) {
	key := *s.apiKey
	key.UserID = userID
	return &key, nil
}
func (s *memoryStore) GetUserByEmail(email string) (*model.User, error) {
	u, ok := s.users[email]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}
func (s *memoryStore) GetAPIKeyByUserID(userID int) (*model.APIKey, error) {
	key := *s.apiKey
	key.UserID = userID
	return &key, nil
}
func (s *memoryStore) GetAPIKey(key string) (*model.APIKey, error) {
	if key != s.apiKey.Key {
		return nil, errors.New("not found")
	}
	return s.apiKey, nil
}
func (s *memoryStore) GetPaymentMethod(code string) (*model.PaymentMethod, error) {
	if !s.methods[code] {
		return nil, errors.New("not found")
	}
	return &model.PaymentMethod{Code: code, IsActive: true}, nil
}
func (s *memoryStore) GetCurrencyRate(code string, btcPrice float64) (*model.CurrencyRate, error) {
	if code != "USD" && code != "BTC" && code != "EUR" {
		return nil, errors.New("not found")
	}
	return &model.CurrencyRate{CurrencyCode: code, RateToUSD: 1, RateToSatoshi: 1}, nil
}
func (s *memoryStore) GetRandomFreeAddress() (string, error) { return "btc-address", nil }
func (s *memoryStore) CreateInvoiceWithDetails(userID int, address string, amount int64, method, currency, promo, ref string, isTest bool) (*model.Invoice, error) {
	i := &model.Invoice{ID: 1, UserID: userID, Address: address, Amount: amount, PaymentMethod: method, Currency: currency, PromoCode: promo, Status: "pending", IsTest: isTest, ExpiresAt: time.Now().Add(time.Hour)}
	s.lastInvoice = i
	return i, nil
}
func (s *memoryStore) AddPaymentEvent(int, string, string) error { return nil }
func (s *memoryStore) GetPromoCode(code string) (*model.PromoCode, error) {
	p, ok := s.promos[code]
	if !ok {
		return nil, errors.New("not found")
	}
	return p, nil
}
func (s *memoryStore) UsePromoCode(code string) error { s.promos[code].UsedCount++; return nil }
func (s *memoryStore) CreatePromoCode(code string, discount float64, maxUses int, expires time.Time) (*model.PromoCode, error) {
	p := &model.PromoCode{ID: len(s.promos) + 1, Code: code, DiscountPercent: discount, MaxUses: maxUses, Active: true, ExpiresAt: expires}
	s.promos[code] = p
	return p, nil
}

func jsonRequest(t *testing.T, method, path string, payload any) *http.Request {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestRegisterAndLogin(t *testing.T) {
	store := newMemoryStore()
	auth := NewAuthHandler(store)

	register := httptest.NewRecorder()
	auth.Register(register, jsonRequest(t, http.MethodPost, "/auth/register", map[string]string{"email": "user@example.com", "password": "secret"}))
	if register.Code != http.StatusCreated {
		t.Fatalf("register status = %d: %s", register.Code, register.Body.String())
	}
	if store.users["user@example.com"].Password == "secret" {
		t.Fatal("password was stored without hashing")
	}

	login := httptest.NewRecorder()
	auth.Login(login, jsonRequest(t, http.MethodPost, "/auth/login", map[string]string{"email": "user@example.com", "password": "secret"}))
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d: %s", login.Code, login.Body.String())
	}

	wrongPassword := httptest.NewRecorder()
	auth.Login(wrongPassword, jsonRequest(t, http.MethodPost, "/auth/login", map[string]string{"email": "user@example.com", "password": "wrong"}))
	if wrongPassword.Code != http.StatusUnauthorized {
		t.Fatalf("wrong-password status = %d", wrongPassword.Code)
	}
}

func TestLoginRejectsUnknownUser(t *testing.T) {
	store := newMemoryStore()
	// Ensure the test does not accidentally depend on bcrypt work in a missing-user path.
	_, _ = bcrypt.GenerateFromPassword([]byte("unused"), bcrypt.MinCost)
	w := httptest.NewRecorder()
	NewAuthHandler(store).Login(w, jsonRequest(t, http.MethodPost, "/auth/login", map[string]string{"email": "missing@example.com", "password": "secret"}))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestCreatePromoCode(t *testing.T) {
	store := newMemoryStore()
	w := httptest.NewRecorder()
	NewPaymentHandler(store).CreatePromoCode(w, jsonRequest(t, http.MethodPost, "/payments/promos/create", map[string]any{"code": "SUMMER2026AA", "discount_percent": 15, "max_uses": 3, "expires_at": time.Now().Add(time.Hour).Format(time.RFC3339)}))
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}
	if store.promos["SUMMER2026AA"].DiscountPercent != 15 {
		t.Fatal("promo code was not saved")
	}
}

func TestCreateInvoiceForEachPaymentMethod(t *testing.T) {
	for _, method := range []string{"bitcoin", "card", "bank"} {
		t.Run(method, func(t *testing.T) {
			store := newMemoryStore()
			h := NewPaymentHandler(store)
			h.btcPrice = func() (float64, error) { return 50000, nil }
			w := httptest.NewRecorder()
			req := jsonRequest(t, http.MethodPost, "/payments/create", map[string]any{"amount": 2500, "payment_method": method, "currency": "USD"})
			req.Header.Set("X-API-Key", "test-key")
			h.CreateInvoice(w, req)
			if w.Code != http.StatusCreated {
				t.Fatalf("status = %d: %s", w.Code, w.Body.String())
			}
			if store.lastInvoice == nil || store.lastInvoice.PaymentMethod != method || store.lastInvoice.Currency != "USD" {
				t.Fatalf("incorrect persisted reference: %#v", store.lastInvoice)
			}
		})
	}
}

func TestCreateInvoiceRejectsUnknownPaymentMethod(t *testing.T) {
	store := newMemoryStore()
	h := NewPaymentHandler(store)
	h.btcPrice = func() (float64, error) { return 50000, nil }
	w := httptest.NewRecorder()
	req := jsonRequest(t, http.MethodPost, "/payments/create", map[string]any{"amount": 1, "payment_method": "cash", "currency": "USD"})
	req.Header.Set("X-API-Key", "test-key")
	h.CreateInvoice(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", w.Code)
	}
}
