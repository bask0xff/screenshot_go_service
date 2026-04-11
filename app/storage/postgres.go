package storage

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
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
		INSERT INTO btc_invoices (user_id, address, amount_usd, amount_btc, status, expires_at)
		VALUES ($1, $2, $3, $4, 'pending', $5)`
	_, err := s.db.Exec(query, userID, address, usdAmount, btcAmount, expiresAt)
	return err
}

func (s *Storage) ConfirmInvoice(address string) (int, float64, error) {
	var userID int
	var amountUSD float64
	query := `
		UPDATE btc_invoices 
		SET status = 'confirmed' 
		WHERE address = $1 AND status = 'pending'
		RETURNING user_id, amount_usd`
	err := s.db.QueryRow(query, address).Scan(&userID, &amountUSD)
	return userID, amountUSD, err
}

// --- HELPERS ---

func generateKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}