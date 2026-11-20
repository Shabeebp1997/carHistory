package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                         uuid.UUID  `json:"id" db:"id"`
	Name                       string     `json:"name" db:"name"`
	Email                      *string    `json:"email" db:"email"`
	PasswordHash               string     `json:"-" db:"password_hash"`
	MobileNumber               *string    `json:"mobile_number" db:"mobile_number"`
	Timezone                   string     `json:"timezone" db:"timezone"`
	EmailNotificationsEnabled  bool       `json:"email_notifications_enabled" db:"email_notifications_enabled"`
	SMSNotificationsEnabled    bool       `json:"sms_notifications_enabled" db:"sms_notifications_enabled"`
	ProfilePicture             *string    `json:"profile_picture" db:"profile_picture"`
	IsAdmin                    bool       `json:"is_admin" db:"is_admin"`
	IsActive                   bool       `json:"is_active" db:"is_active"`
	CreatedAt                  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt                  time.Time  `json:"updated_at" db:"updated_at"`
}

type PasswordResetToken struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	Token     string     `db:"token"`
	ExpiresAt time.Time  `db:"expires_at"`
	UsedAt    *time.Time `db:"used_at"`
	CreatedAt time.Time  `db:"created_at"`
}

type RefreshToken struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	Token     string     `db:"token"`
	ExpiresAt time.Time  `db:"expires_at"`
	RevokedAt *time.Time `db:"revoked_at"`
	CreatedAt time.Time  `db:"created_at"`
}

type RegisterRequest struct {
	Name         string  `json:"name" validate:"required,min=2,max=255"`
	Email        *string `json:"email" validate:"omitempty,email"`
	Password     string  `json:"password" validate:"required,min=8"`
	MobileNumber *string `json:"mobile_number"`
	IsAdmin      bool    `json:"is_admin"`
}

type LoginRequest struct {
	Email    *string `json:"email" validate:"omitempty,email"`
	Phone    *string `json:"phone"`
	Password string  `json:"password" validate:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

type UpdateProfileRequest struct {
	Name                       string  `json:"name" validate:"omitempty,min=2,max=255"`
	MobileNumber               *string `json:"mobile_number"`
	Timezone                   string  `json:"timezone"`
	EmailNotificationsEnabled  *bool   `json:"email_notifications_enabled"`
	SMSNotificationsEnabled    *bool   `json:"sms_notifications_enabled"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *User  `json:"user"`
}
