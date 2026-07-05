package model

import "time"

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type APIKey struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Key       string    `json:"key"`
	Tier      string    `json:"tier"`
	Requests  int       `json:"requests"`
	CreatedAt time.Time `json:"created_at"`
}

type Invoice struct {
	ID            int        `json:"id"`
	UserID        int        `json:"user_id"`
	Address       string     `json:"address"`
	AmountUSD     float64    `json:"amount_usd"`
	AmountBTC     float64    `json:"amount_btc"`
	Status        string     `json:"status"`
	PaymentMethod string     `json:"payment_method"`
	Currency      string     `json:"currency"`
	PromoCode     string     `json:"promo_code,omitempty"`
	PaymentRef    string     `json:"payment_reference,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	ExpiresAt     time.Time  `json:"expires_at"`
	ConfirmedAt   *time.Time `json:"confirmed_at,omitempty"`
	CancelledAt   *time.Time `json:"cancelled_at,omitempty"`
}

type PromoCode struct {
	ID              int       `json:"id"`
	Code            string    `json:"code"`
	DiscountPercent float64   `json:"discount_percent"`
	MaxUses         int       `json:"max_uses"`
	UsedCount       int       `json:"used_count"`
	Active          bool      `json:"active"`
	ExpiresAt       time.Time `json:"expires_at"`
	CreatedAt       time.Time `json:"created_at"`
}

type PaymentEvent struct {
	ID        int       `json:"id"`
	InvoiceID int       `json:"invoice_id"`
	EventType string    `json:"event_type"`
	Payload   string    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}
