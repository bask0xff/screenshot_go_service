CREATE TABLE api_keys (
                          id         SERIAL PRIMARY KEY,
                          user_id    INTEGER REFERENCES users(id) ON DELETE CASCADE,
                          key        VARCHAR(64) UNIQUE NOT NULL,
                          tier       VARCHAR(20) DEFAULT 'free',
                          requests   INTEGER DEFAULT 0,
                          created_at TIMESTAMP DEFAULT NOW()
);