package storage

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"screenshot-api/config"
	"screenshot-api/model"
)

type Storage struct {
	db *sql.DB
}

func New(cfg *config.Config) (*Storage, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) RunMigrations(cfg *config.Config) error {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

// --- USER METHODS ---

func (s *Storage) CreateUser(email, passwordHash string) (*model.User, error) {
	user := &model.User{}
	err := s.db.QueryRow(
		`INSERT INTO users (email, password) VALUES ($1, $2)
         RETURNING id, email, created_at`,
		email, passwordHash,
	).Scan(&user.ID, &user.Email, &user.CreatedAt)
	return user, err
}

func (s *Storage) GetUserByEmail(email string) (*model.User, error) {
	user := &model.User{}
	err := s.db.QueryRow(
		`SELECT id, email, password, created_at FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt)
	return user, err
}

func (s *Storage) GetUserBalance(userID int) (float64, error) {
	var balance float64
	err := s.db.QueryRow(`SELECT balance_usd FROM users WHERE id = $1`, userID).Scan(&balance)
	return balance, err
}

func (s *Storage) UpdateUserBalance(userID int, amount float64) error {
	_, err := s.db.Exec(`UPDATE users SET balance_usd = balance_usd + $1 WHERE id = $2`, amount, userID)
	return err
}

// --- API KEY METHODS ---

func (s *Storage) CreateAPIKey(userID int) (*model.APIKey, error) {
	key, err := generateKey()
	if err != nil {
		return nil, err
	}

	apiKey := &model.APIKey{}
	err = s.db.QueryRow(
		`INSERT INTO api_keys (user_id, key) VALUES ($1, $2)
         RETURNING id, user_id, key, tier, requests, created_at`,
		userID, key,
	).Scan(&apiKey.ID, &apiKey.UserID, &apiKey.Key, &apiKey.Tier, &apiKey.Requests, &apiKey.CreatedAt)
	return apiKey, err
}

func (s *Storage) GetAPIKey(key string) (*model.APIKey, error) {
	apiKey := &model.APIKey{}
	err := s.db.QueryRow(
		`SELECT id, user_id, key, tier, requests, created_at FROM api_keys WHERE key = $1`,
		key,
	).Scan(&apiKey.ID, &apiKey.UserID, &apiKey.Key, &apiKey.Tier, &apiKey.Requests, &apiKey.CreatedAt)
	return apiKey, err
}

func (s *Storage) GetAPIKeyByUserID(userID int) (*model.APIKey, error) {
	apiKey := &model.APIKey{}
	err := s.db.QueryRow(
		`SELECT id, user_id, key, tier, requests, created_at FROM api_keys WHERE user_id = $1`,
		userID,
	).Scan(&apiKey.ID, &apiKey.UserID, &apiKey.Key, &apiKey.Tier, &apiKey.Requests, &apiKey.CreatedAt)
	return apiKey, err
}

func (s *Storage) IncrementRequests(key string) error {
	_, err := s.db.Exec(
		`UPDATE api_keys SET requests = requests + 1 WHERE key = $1`, key)
	return err
}

// --- PAYMENT & INVOICE METHODS ---

func (s *Storage) GetRandomFreeAddress() (string, error) {
	var addr string
	// Выбираем адрес из btcaddress2, которого нет в активных или завершенных инвойсах
	query := `
		SELECT address FROM btcaddress2 
		WHERE address NOT IN (SELECT address FROM btc_invoices WHERE status != 'expired') 
		ORDER BY RANDOM() LIMIT 1`
	err := s.db.QueryRow(query).Scan(&addr)
	if err != nil {
		return "", err
	}
	return addr, nil
}

func (s *Storage) CreateInvoice(userID int, address string, usdAmount, btcAmount float64) error {
	expiresAt := time.Now().Add(3 * time.Hour)
	query := `
		INSERT INTO btc_invoices (user_id, address, amount_usd, amount_btc, amount_satoshi, status, expires_at)
		VALUES ($1, $2, $3, $4, 0, 'pending', $5)`
	_, err := s.db.Exec(query, userID, address, usdAmount, btcAmount, expiresAt)
	return err
}

func (s *Storage) ConfirmInvoice(address string) (int, int64, error) {
	var userID int
	var amountSatoshi int64
	query := `
		UPDATE btc_invoices 
		SET status = 'confirmed', confirmed_at = NOW()
		WHERE address = $1 AND status = 'pending'
		RETURNING user_id, amount_satoshi`
	err := s.db.QueryRow(query, address).Scan(&userID, &amountSatoshi)
	return userID, amountSatoshi, err
}

func (s *Storage) CreateInvoiceWithDetails(userID int, address string, amountSatoshi int64, paymentMethod, currency, promoCode, paymentRef string, isTest bool) (*model.Invoice, error) {
	invoice := &model.Invoice{}
	expiresAt := time.Now().Add(3 * time.Hour)
	err := s.db.QueryRow(`
		INSERT INTO btc_invoices (user_id, address, amount_satoshi, status, payment_method, currency, promo_code, payment_reference, is_test, expires_at)
		VALUES ($1, $2, $3, 'pending', $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, address, amount_satoshi, status, payment_method, currency, promo_code, payment_reference, is_test, created_at, expires_at
	`, userID, address, amountSatoshi, paymentMethod, currency, promoCode, paymentRef, isTest, expiresAt).Scan(
		&invoice.ID,
		&invoice.UserID,
		&invoice.Address,
		&invoice.Amount,
		&invoice.Status,
		&invoice.PaymentMethod,
		&invoice.Currency,
		&invoice.PromoCode,
		&invoice.PaymentRef,
		&invoice.IsTest,
		&invoice.CreatedAt,
		&invoice.ExpiresAt,
	)
	return invoice, err
}

func (s *Storage) GetInvoiceByAddress(address string) (*model.Invoice, error) {
	invoice := &model.Invoice{}
	err := s.db.QueryRow(`
		SELECT id, user_id, address, amount_satoshi, status, payment_method, currency, promo_code, payment_reference, is_test, created_at, expires_at, confirmed_at, cancelled_at
		FROM btc_invoices WHERE address = $1
	`, address).Scan(
		&invoice.ID,
		&invoice.UserID,
		&invoice.Address,
		&invoice.Amount,
		&invoice.Status,
		&invoice.PaymentMethod,
		&invoice.Currency,
		&invoice.PromoCode,
		&invoice.PaymentRef,
		&invoice.IsTest,
		&invoice.CreatedAt,
		&invoice.ExpiresAt,
		&invoice.ConfirmedAt,
		&invoice.CancelledAt,
	)
	return invoice, err
}

func (s *Storage) CancelInvoice(address string) error {
	_, err := s.db.Exec(`
		UPDATE btc_invoices SET status = 'cancelled', cancelled_at = NOW() WHERE address = $1 AND status = 'pending'
	`, address)
	return err
}

func (s *Storage) ListPendingInvoices() ([]*model.Invoice, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, address, amount_satoshi, status, payment_method, currency, promo_code, payment_reference, is_test, created_at, expires_at, confirmed_at, cancelled_at
		FROM btc_invoices WHERE status = 'pending' ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []*model.Invoice
	for rows.Next() {
		invoice := &model.Invoice{}
		err := rows.Scan(
			&invoice.ID,
			&invoice.UserID,
			&invoice.Address,
			&invoice.Amount,
			&invoice.Status,
			&invoice.PaymentMethod,
			&invoice.Currency,
			&invoice.PromoCode,
			&invoice.PaymentRef,
			&invoice.IsTest,
			&invoice.CreatedAt,
			&invoice.ExpiresAt,
			&invoice.ConfirmedAt,
			&invoice.CancelledAt,
		)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}
	return invoices, nil
}

func (s *Storage) GetCurrencyRate(code string, btcPrice float64) (*model.CurrencyRate, error) {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	if normalized == "" {
		normalized = "USD"
	}

	if normalized == "BTC" || normalized == "XBT" {
		return &model.CurrencyRate{CurrencyCode: normalized, RateToUSD: btcPrice, RateToSatoshi: 100000000}, nil
	}
	if normalized == "USD" || normalized == "USDT" {
		return &model.CurrencyRate{CurrencyCode: normalized, RateToUSD: 1, RateToSatoshi: int64(100000000 / btcPrice)}, nil
	}

	var rate model.CurrencyRate
	err := s.db.QueryRow(`
		SELECT currency_code, rate_to_usd, rate_to_satoshi
		FROM currency_rates
		WHERE currency_code = $1
		ORDER BY effective_at DESC, created_at DESC
		LIMIT 1
	`, normalized).Scan(&rate.CurrencyCode, &rate.RateToUSD, &rate.RateToSatoshi)
	if err == nil {
		return &rate, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	if _, err := s.db.Exec(`INSERT INTO currencies (code, name, is_crypto, is_active) VALUES ($1, $2, false, true) ON CONFLICT (code) DO NOTHING`, normalized, normalized); err != nil {
		return nil, err
	}
	if _, err := s.db.Exec(`INSERT INTO currency_rates (currency_code, rate_to_usd, rate_to_satoshi, effective_at) SELECT $1, 1.0, 0, NOW() WHERE NOT EXISTS (SELECT 1 FROM currency_rates WHERE currency_code = $1)`, normalized); err != nil {
		return nil, err
	}

	return &model.CurrencyRate{CurrencyCode: normalized, RateToUSD: 1, RateToSatoshi: 0}, nil
}

func (s *Storage) CreatePromoCode(code string, discountPercent float64, maxUses int, expiresAt time.Time) (*model.PromoCode, error) {
	promo := &model.PromoCode{}
	err := s.db.QueryRow(`
		INSERT INTO promo_codes (code, discount_percent, max_uses, used_count, active, expires_at)
		VALUES ($1, $2, $3, 0, true, $4)
		RETURNING id, code, discount_percent, max_uses, used_count, active, expires_at, created_at
	`, code, discountPercent, maxUses, expiresAt).Scan(
		&promo.ID,
		&promo.Code,
		&promo.DiscountPercent,
		&promo.MaxUses,
		&promo.UsedCount,
		&promo.Active,
		&promo.ExpiresAt,
		&promo.CreatedAt,
	)
	return promo, err
}

func (s *Storage) GetPromoCode(code string) (*model.PromoCode, error) {
	promo := &model.PromoCode{}
	err := s.db.QueryRow(`
		SELECT id, code, discount_percent, max_uses, used_count, active, expires_at, created_at
		FROM promo_codes WHERE code = $1
	`, code).Scan(
		&promo.ID,
		&promo.Code,
		&promo.DiscountPercent,
		&promo.MaxUses,
		&promo.UsedCount,
		&promo.Active,
		&promo.ExpiresAt,
		&promo.CreatedAt,
	)
	return promo, err
}

func (s *Storage) UsePromoCode(code string) error {
	_, err := s.db.Exec(`
		UPDATE promo_codes SET used_count = used_count + 1 WHERE code = $1
	`, code)
	return err
}

func (s *Storage) AddPaymentEvent(invoiceID int, eventType, payload string) error {
	_, err := s.db.Exec(`
		INSERT INTO payment_events (invoice_id, event_type, payload) VALUES ($1, $2, $3)
	`, invoiceID, eventType, payload)
	return err
}

// --- HELPERS ---

func generateKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
