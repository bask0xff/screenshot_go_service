ALTER TABLE btc_invoices DROP COLUMN IF EXISTS amount_satoshi;
DROP TABLE IF EXISTS currency_rates;
DROP TABLE IF EXISTS currencies;
