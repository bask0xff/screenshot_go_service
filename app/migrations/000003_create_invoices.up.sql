CREATE TABLE IF NOT EXISTS btc_invoices (
    id                SERIAL PRIMARY KEY,
    user_id           INTEGER REFERENCES users(id) ON DELETE CASCADE,
    address           VARCHAR(255) UNIQUE NOT NULL,
    amount_usd        DECIMAL(10, 2) DEFAULT 0.00,
    amount_btc        DECIMAL(16, 8) DEFAULT 0.00000000,
    amount_satoshi    BIGINT DEFAULT 0,
    status            VARCHAR(20) DEFAULT 'pending',
    payment_method    VARCHAR(50) DEFAULT 'bitcoin',
    currency          VARCHAR(10) DEFAULT 'USD',
    promo_code        VARCHAR(100),
    payment_reference VARCHAR(255),
    is_test           BOOLEAN DEFAULT FALSE,
    created_at        TIMESTAMP DEFAULT NOW(),
    expires_at        TIMESTAMP NOT NULL,
    confirmed_at     TIMESTAMP,
    cancelled_at      TIMESTAMP
    );

ALTER TABLE users ADD COLUMN IF NOT EXISTS balance_usd DECIMAL(10, 2) DEFAULT 0.00;