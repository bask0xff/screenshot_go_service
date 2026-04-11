CREATE TABLE btc_invoices (
                              id          SERIAL PRIMARY KEY,
                              user_id     INTEGER REFERENCES users(id) ON DELETE CASCADE,
                              address     VARCHAR(255) UNIQUE NOT NULL,
                              amount_usd  DECIMAL(10, 2) NOT NULL, -- Сколько баксов он хочет зачислить
                              amount_btc  DECIMAL(16, 8) NOT NULL, -- Сколько сатоши мы ждем
                              status      VARCHAR(20) DEFAULT 'pending', -- pending, confirmed, expired
                              created_at  TIMESTAMP DEFAULT NOW()
                              expires_at  TIMESTAMP NOT NULL

                          );

ALTER TABLE users ADD COLUMN balance_usd DECIMAL(10, 2) DEFAULT 0.00;