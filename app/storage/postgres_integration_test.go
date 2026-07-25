//go:build integration

package storage

import (
	"database/sql"
	"strings"
	"testing"

	"screenshot-api/config"
)

func TestMigrationsNormalizeInvoicePaymentReferences(t *testing.T) {
	cfg := &config.Config{
		DBHost:     "postgres-integration",
		DBPort:     "5432",
		DBUser:     "integration",
		DBPassword: "integration",
		DBName:     "screenshot_integration",
	}

	store, err := New(cfg)
	if err != nil {
		t.Fatalf("connect to isolated PostgreSQL: %v", err)
	}
	t.Cleanup(func() { _ = store.db.Close() })

	if err := store.RunMigrations(cfg); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	assertColumnExists(t, store.db, "btc_invoices", "payment_method_id")
	assertColumnExists(t, store.db, "btc_invoices", "currency_id")
	assertColumnMissing(t, store.db, "btc_invoices", "payment_method")
	assertColumnMissing(t, store.db, "btc_invoices", "currency")
	assertForeignKey(t, store.db, "btc_invoices_payment_method_id_fkey", "payment_methods")
	assertForeignKey(t, store.db, "btc_invoices_currency_id_fkey", "currencies")

	for _, code := range []string{"bitcoin", "card", "bank"} {
		if _, err := store.GetPaymentMethod(code); err != nil {
			t.Fatalf("seeded payment method %q is unavailable: %v", code, err)
		}
	}

	user, err := store.CreateUser("integration@example.com", "password-hash")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	invoice, err := store.CreateInvoiceWithDetails(user.ID, "integration-address", 2500, "card", "USD", "", "", false)
	if err != nil {
		t.Fatalf("create invoice with reference keys: %v", err)
	}
	if invoice.PaymentMethod != "card" || invoice.Currency != "USD" {
		t.Fatalf("unexpected invoice references: method=%q currency=%q", invoice.PaymentMethod, invoice.Currency)
	}

	_, err = store.db.Exec(`
		INSERT INTO btc_invoices (user_id, address, expires_at, payment_method_id, currency_id)
		VALUES ($1, 'invalid-reference', NOW(), 999999, 999999)
	`, user.ID)
	if err == nil {
		t.Fatal("invoice with nonexistent reference IDs was accepted")
	}
}

func assertColumnExists(t *testing.T, db *sql.DB, table, column string) {
	t.Helper()
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2
		)
	`, table, column).Scan(&exists)
	if err != nil || !exists {
		t.Fatalf("expected column %s.%s to exist: %v", table, column, err)
	}
}

func assertColumnMissing(t *testing.T, db *sql.DB, table, column string) {
	t.Helper()
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2
		)
	`, table, column).Scan(&exists)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatalf("expected legacy column %s.%s to be removed", table, column)
	}
}

func assertForeignKey(t *testing.T, db *sql.DB, constraint, referencedTable string) {
	t.Helper()
	var table string
	err := db.QueryRow(`
		SELECT ccu.table_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.constraint_column_usage ccu
		  ON ccu.constraint_name = tc.constraint_name AND ccu.table_schema = tc.table_schema
		WHERE tc.table_schema = 'public' AND tc.constraint_name = $1 AND tc.constraint_type = 'FOREIGN KEY'
	`, constraint).Scan(&table)
	if err != nil || !strings.EqualFold(table, referencedTable) {
		t.Fatalf("expected foreign key %s to reference %s; got %q, err=%v", constraint, referencedTable, table, err)
	}
}
