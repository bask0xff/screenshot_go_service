CREATE TABLE users (
    id         SERIAL PRIMARY KEY,
    email      VARCHAR(255) UNIQUE NOT NULL,
    password   VARCHAR(255) NOT NULL,
    balance_usd DECIMAL(10, 2) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT NOW()
);
