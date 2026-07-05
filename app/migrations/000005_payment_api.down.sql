DROP TABLE IF EXISTS payment_events;
DROP TABLE IF EXISTS promo_codes;

ALTER TABLE btc_invoices
    DROP COLUMN IF EXISTS payment_method,
    DROP COLUMN IF EXISTS currency,
    DROP COLUMN IF EXISTS promo_code,
    DROP COLUMN IF EXISTS payment_reference,
    DROP COLUMN IF EXISTS confirmed_at,
    DROP COLUMN IF EXISTS cancelled_at;
