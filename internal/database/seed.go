package database

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// SeedAdmin ensures the configured ADMIN_EMAIL user exists and has is_admin=TRUE.
//
// Behaviour:
//   - ADMIN_EMAIL set, user already exists → promotes to admin (idempotent)
//   - ADMIN_EMAIL + ADMIN_PASSWORD set, user does NOT exist → creates the admin account
//   - ADMIN_EMAIL set but user missing and no password given → logs a hint
//   - ADMIN_EMAIL empty → no-op
func SeedAdmin(db *sqlx.DB, email, password string) error {
	if email == "" {
		return nil
	}

	ctx := context.Background()

	var count int
	if err := db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM users WHERE email = $1`, email); err != nil {
		return err
	}

	if count > 0 {
		_, err := db.ExecContext(ctx,
			`UPDATE users SET is_admin = TRUE, updated_at = $1 WHERE email = $2`,
			time.Now(), email)
		if err != nil {
			return err
		}
		log.Printf("[Admin] User %s promoted to admin", email)
		return nil
	}

	// User does not exist — create if password supplied
	if password == "" {
		log.Printf("[Admin] No user found with email %s. "+
			"Register that account first, OR also set ADMIN_PASSWORD in .env to auto-create it.", email)
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = db.ExecContext(ctx, `
		INSERT INTO users
			(id, name, email, password_hash, timezone, is_admin, is_active,
			 email_notifications_enabled, sms_notifications_enabled, created_at, updated_at)
		VALUES ($1, 'Admin', $2, $3, 'UTC', TRUE, TRUE, TRUE, FALSE, $4, $4)
		ON CONFLICT (email) DO UPDATE SET is_admin = TRUE, updated_at = $4`,
		uuid.New(), email, string(hash), now)
	if err != nil {
		return err
	}
	log.Printf("[Admin] Admin account created: %s", email)
	return nil
}
