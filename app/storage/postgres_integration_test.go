//go:build integration

package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"screenshot-api/config"
)

func TestMigrationsNormalizeInvoicePaymentReferencesAndPromoCodes(t *testing.T) {
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

	uniqueID := time.Now().UnixNano()
	user, err := store.CreateUser(fmt.Sprintf("integration-%d@example.com", uniqueID), "password-hash")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	for index, paymentMethod := range []string{"bitcoin", "card", "bank"} {
		promoCode := fmt.Sprintf("INTG%07d%d", uniqueID%10000000, index)
		promo, err := store.CreatePromoCode(promoCode, 15, 3, time.Now().Add(time.Hour))
		if err != nil {
			t.Fatalf("create promo code for %s payment: %v", paymentMethod, err)
		}
		if promo.Code != promoCode || promo.DiscountPercent != 15 || promo.MaxUses != 3 || promo.UsedCount != 0 {
			t.Fatalf("unexpected created promo code for %s payment: %#v", paymentMethod, promo)
		}
		if err := store.UsePromoCode(promoCode); err != nil {
			t.Fatalf("use promo code for %s payment: %v", paymentMethod, err)
		}
		promo, err = store.GetPromoCode(promoCode)
		if err != nil {
			t.Fatalf("get promo code for %s payment: %v", paymentMethod, err)
		}
		if promo.UsedCount != 1 {
			t.Fatalf("expected promo code usage count 1 for %s payment, got %d", paymentMethod, promo.UsedCount)
		}

		invoice, err := store.CreateInvoiceWithDetails(
			user.ID,
			fmt.Sprintf("integration-address-%d-%d", uniqueID, index),
			2500,
			paymentMethod,
			"USD",
			promoCode,
			"",
			false,
		)
		if err != nil {
			t.Fatalf("create %s invoice with promo code: %v", paymentMethod, err)
		}
		if invoice.PaymentMethod != paymentMethod || invoice.Currency != "USD" || invoice.PromoCode != promoCode {
			t.Fatalf("unexpected %s invoice data: method=%q currency=%q promo=%q", paymentMethod, invoice.PaymentMethod, invoice.Currency, invoice.PromoCode)
		}
	}

	_, err = store.db.Exec(`
		INSERT INTO btc_invoices (user_id, address, expires_at, payment_method_id, currency_id)
		VALUES ($1, $2, NOW(), 999999, 999999)
	`, user.ID, fmt.Sprintf("invalid-reference-%d", uniqueID))
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
