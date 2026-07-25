ALTER TABLE btc_invoices
    ADD COLUMN payment_method VARCHAR(32),
    ADD COLUMN currency VARCHAR(10);

UPDATE btc_invoices AS i
SET payment_method = pm.code
FROM payment_methods AS pm
WHERE pm.id = i.payment_method_id;

UPDATE btc_invoices AS i
SET currency = c.code
FROM currencies AS c
WHERE c.id = i.currency_id;

ALTER TABLE btc_invoices
    ALTER COLUMN payment_method SET NOT NULL,
    ALTER COLUMN currency SET NOT NULL,
    DROP CONSTRAINT IF EXISTS btc_invoices_payment_method_id_fkey,
    DROP CONSTRAINT IF EXISTS btc_invoices_currency_id_fkey,
    DROP COLUMN payment_method_id,
    DROP COLUMN currency_id;

DROP TABLE IF EXISTS payment_methods;
