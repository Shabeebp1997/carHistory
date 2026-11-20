package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"carhistory/internal/models"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, name, email, password_hash, mobile_number, timezone,
		                   email_notifications_enabled, sms_notifications_enabled,
		                   is_admin, is_active, created_at, updated_at)
		VALUES (:id, :name, :email, :password_hash, :mobile_number, :timezone,
		        :email_notifications_enabled, :sms_notifications_enabled,
		        :is_admin, :is_active, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.GetContext(ctx, &user, `SELECT * FROM users WHERE email = $1 AND is_active = true`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*models.User, error) {
	var user models.User
	err := r.db.GetContext(ctx, &user, `SELECT * FROM users WHERE mobile_number = $1 AND is_active = true`, phone)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.GetContext(ctx, &user, `SELECT * FROM users WHERE id = $1 AND is_active = true`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users SET
			name = :name,
			mobile_number = :mobile_number,
			timezone = :timezone,
			email_notifications_enabled = :email_notifications_enabled,
			sms_notifications_enabled = :sms_notifications_enabled,
			updated_at = :updated_at
		WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}

func (r *UserRepository) UpdateProfilePicture(ctx context.Context, userID uuid.UUID, url string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET profile_picture = $1, updated_at = $2 WHERE id = $3`,
		url, time.Now(), userID)
	return err
}

func (r *UserRepository) Deactivate(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET is_active = FALSE, updated_at = $1 WHERE id = $2`,
		time.Now(), userID)
	return err
}

func (r *UserRepository) SetAdminStatus(ctx context.Context, userID uuid.UUID, isAdmin bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET is_admin = $1, updated_at = $2 WHERE id = $3`,
		isAdmin, time.Now(), userID)
	return err
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`,
		passwordHash, time.Now(), userID)
	return err
}

func (r *UserRepository) CreatePasswordResetToken(ctx context.Context, token *models.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (id, user_id, token, expires_at, created_at)
		VALUES (:id, :user_id, :token, :expires_at, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, token)
	return err
}

func (r *UserRepository) FindPasswordResetToken(ctx context.Context, token string) (*models.PasswordResetToken, error) {
	var t models.PasswordResetToken
	err := r.db.GetContext(ctx, &t,
		`SELECT * FROM password_reset_tokens WHERE token = $1 AND used_at IS NULL AND expires_at > NOW()`,
		token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &t, err
}

func (r *UserRepository) MarkPasswordResetTokenUsed(ctx context.Context, tokenID uuid.UUID) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE password_reset_tokens SET used_at = $1 WHERE id = $2`,
		now, tokenID)
	return err
}

func (r *UserRepository) CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at)
		VALUES (:id, :user_id, :token, :expires_at, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, token)
	return err
}

func (r *UserRepository) FindRefreshToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	var t models.RefreshToken
	err := r.db.GetContext(ctx, &t,
		`SELECT * FROM refresh_tokens WHERE token = $1 AND revoked_at IS NULL AND expires_at > NOW()`,
		token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &t, err
}

func (r *UserRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked_at = $1 WHERE token = $2`,
		now, token)
	return err
}

func (r *UserRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL`,
		now, userID)
	return err
}
