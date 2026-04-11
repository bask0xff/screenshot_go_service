CREATE TABLE btc_invoices (
                              id         SERIAL PRIMARY KEY,
                              user_id    INTEGER REFERENCES users(id) ON DELETE CASCADE,
                              address    VARCHAR(255) UNIQUE NOT NULL,
                              amount_usd DECIMAL(10, 2) NOT NULL,
                              amount_btc DECIMAL(16, 8) NOT NULL,
                              status     VARCHAR(20) DEFAULT 'pending',
                              created_at TIMESTAMP DEFAULT NOW(),
                              expires_at TIMESTAMP NOT NULL
);