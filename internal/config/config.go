package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	DB       DBConfig
	JWT      JWTConfig
	SMTP     SMTPConfig
	SMS      SMSConfig
	Storage  StorageConfig
	App      AppConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DBConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	Name         string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

type JWTConfig struct {
	Secret                    string
	AccessTokenExpiryMinutes  int
	RefreshTokenExpiryDays    int
}

type SMTPConfig struct {
	Host         string
	Port         int
	Username     string
	Password     string
	FromEmail    string
	FromName     string
}

type SMSConfig struct {
	Provider    string
	AccountSID  string
	AuthToken   string
	FromNumber  string
}

type StorageConfig struct {
	Type      string
	LocalPath string
	BaseURL   string
}

type AppConfig struct {
	BaseURL                    string
	InvitationExpiryHours      int
	PasswordResetExpiryMinutes int
	ReminderCronSchedule       string
	AdminEmail                 string
	AdminPassword              string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		DB: DBConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnv("DB_PORT", "5432"),
			User:         getEnv("DB_USER", "postgres"),
			Password:     getEnv("DB_PASSWORD", ""),
			Name:         getEnv("DB_NAME", "carhistory"),
			SSLMode:      getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 5),
		},
		JWT: JWTConfig{
			Secret:                   getEnv("JWT_SECRET", "change_me_in_production"),
			AccessTokenExpiryMinutes: getEnvInt("JWT_ACCESS_TOKEN_EXPIRY_MINUTES", 15),
			RefreshTokenExpiryDays:   getEnvInt("JWT_REFRESH_TOKEN_EXPIRY_DAYS", 7),
		},
		SMTP: SMTPConfig{
			Host:      getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:      getEnvInt("SMTP_PORT", 587),
			Username:  getEnv("SMTP_USERNAME", ""),
			Password:  getEnv("SMTP_PASSWORD", ""),
			FromEmail: getEnv("EMAIL_FROM", "noreply@carhistory.app"),
			FromName:  getEnv("EMAIL_FROM_NAME", "Car History"),
		},
		SMS: SMSConfig{
			Provider:   getEnv("SMS_PROVIDER", "twilio"),
			AccountSID: getEnv("SMS_ACCOUNT_SID", ""),
			AuthToken:  getEnv("SMS_AUTH_TOKEN", ""),
			FromNumber: getEnv("SMS_FROM_NUMBER", ""),
		},
		Storage: StorageConfig{
			Type:      getEnv("STORAGE_TYPE", "local"),
			LocalPath: getEnv("STORAGE_LOCAL_PATH", "./uploads"),
			BaseURL:   getEnv("STORAGE_BASE_URL", "http://localhost:8080/uploads"),
		},
		App: AppConfig{
			BaseURL:                   getEnv("APP_BASE_URL", "http://localhost:3000"),
			InvitationExpiryHours:     getEnvInt("INVITATION_EXPIRY_HOURS", 72),
			PasswordResetExpiryMinutes: getEnvInt("PASSWORD_RESET_EXPIRY_MINUTES", 30),
			ReminderCronSchedule:      getEnv("REMINDER_CRON_SCHEDULE", "0 2 * * *"),
			AdminEmail:                 getEnv("ADMIN_EMAIL", ""),
			AdminPassword:              getEnv("ADMIN_PASSWORD", ""),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
