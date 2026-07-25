-- Reference data used by invoices.  Codes are stable public API values; invoice
-- rows reference the corresponding integer keys to preserve referential integrity.
CREATE TABLE IF NOT EXISTS payment_methods (
    id SERIAL PRIMARY KEY,
    code VARCHAR(32) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO payment_methods (code, name) VALUES
    ('bitcoin', 'Bitcoin'),
    ('card', 'Bank card'),
    ('bank', 'Bank transfer')
ON CONFLICT (code) DO NOTHING;

INSERT INTO currencies (code, name, is_crypto, is_active) VALUES
    ('USD', 'US Dollar', FALSE, TRUE),
    ('EUR', 'Euro', FALSE, TRUE),
    ('BTC', 'Bitcoin', TRUE, TRUE)
ON CONFLICT (code) DO NOTHING;

-- Preserve values introduced by older application versions before removing the
-- denormalized columns. Unknown historical codes remain valid reference data.
INSERT INTO payment_methods (code, name)
SELECT DISTINCT lower(trim(payment_method)), payment_method
FROM btc_invoices
WHERE payment_method IS NOT NULL AND trim(payment_method) <> ''
ON CONFLICT (code) DO NOTHING;

INSERT INTO currencies (code, name, is_crypto, is_active)
SELECT DISTINCT upper(trim(currency)), upper(trim(currency)), FALSE, TRUE
FROM btc_invoices
WHERE currency IS NOT NULL AND trim(currency) <> ''
ON CONFLICT (code) DO NOTHING;

ALTER TABLE btc_invoices
    ADD COLUMN IF NOT EXISTS payment_method_id INTEGER,
    ADD COLUMN IF NOT EXISTS currency_id INTEGER;

UPDATE btc_invoices AS i
SET payment_method_id = pm.id
FROM payment_methods AS pm
WHERE pm.code = lower(trim(i.payment_method));

UPDATE btc_invoices AS i
SET currency_id = c.id
FROM currencies AS c
WHERE c.code = upper(trim(i.currency));

ALTER TABLE btc_invoices
    ALTER COLUMN payment_method_id SET NOT NULL,
    ALTER COLUMN currency_id SET NOT NULL,
    ADD CONSTRAINT btc_invoices_payment_method_id_fkey
        FOREIGN KEY (payment_method_id) REFERENCES payment_methods(id),
    ADD CONSTRAINT btc_invoices_currency_id_fkey
        FOREIGN KEY (currency_id) REFERENCES currencies(id),
    DROP COLUMN payment_method,
    DROP COLUMN currency;

CREATE INDEX IF NOT EXISTS btc_invoices_payment_method_id_idx ON btc_invoices(payment_method_id);
CREATE INDEX IF NOT EXISTS btc_invoices_currency_id_idx ON btc_invoices(currency_id);
