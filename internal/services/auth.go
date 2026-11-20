package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	jwtpkg "carhistory/pkg/jwt"
	"carhistory/internal/config"
	"carhistory/internal/models"
	"carhistory/internal/repository"
	"carhistory/pkg/utils"
)

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrEmailTaken          = errors.New("email already in use")
	ErrPhoneTaken          = errors.New("phone number already in use")
	ErrContactRequired     = errors.New("email or phone number is required")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidToken        = errors.New("invalid or expired token")
)

type AuthService struct {
	userRepo    *repository.UserRepository
	jwtManager  *jwtpkg.Manager
	cfg         *config.Config
	emailSvc    *EmailService
}

func NewAuthService(
	userRepo *repository.UserRepository,
	jwtManager *jwtpkg.Manager,
	cfg *config.Config,
	emailSvc *EmailService,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		cfg:        cfg,
		emailSvc:   emailSvc,
	}
}

func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
	if req.Email == nil && req.MobileNumber == nil {
		return nil, ErrContactRequired
	}

	if req.Email != nil {
		existing, err := s.userRepo.FindByEmail(ctx, *req.Email)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, ErrEmailTaken
		}
	}

	if req.MobileNumber != nil {
		existing, err := s.userRepo.FindByPhone(ctx, *req.MobileNumber)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, ErrPhoneTaken
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &models.User{
		ID:                        uuid.New(),
		Name:                      req.Name,
		Email:                     req.Email,
		PasswordHash:              string(hash),
		MobileNumber:              req.MobileNumber,
		Timezone:                  "UTC",
		EmailNotificationsEnabled: req.Email != nil,
		SMSNotificationsEnabled:   req.MobileNumber != nil && req.Email == nil,
		IsAdmin:                   req.IsAdmin,
		IsActive:                  true,
		CreatedAt:                 now,
		UpdatedAt:                 now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
	if req.Email == nil && req.Phone == nil {
		return nil, ErrContactRequired
	}

	var user *models.User
	var err error

	if req.Email != nil {
		user, err = s.userRepo.FindByEmail(ctx, *req.Email)
	} else {
		user, err = s.userRepo.FindByPhone(ctx, *req.Phone)
	}

	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.userRepo.RevokeRefreshToken(ctx, refreshToken)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (*models.AuthResponse, error) {
	storedToken, err := s.userRepo.FindRefreshToken(ctx, refreshTokenStr)
	if err != nil {
		return nil, err
	}
	if storedToken == nil {
		return nil, ErrInvalidToken
	}

	claims, err := s.jwtManager.ValidateToken(refreshTokenStr)
	if err != nil {
		return nil, ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if err := s.userRepo.RevokeRefreshToken(ctx, refreshTokenStr); err != nil {
		return nil, err
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}

	token, err := utils.GenerateSecureToken(32)
	if err != nil {
		return err
	}

	now := time.Now()
	resetToken := &models.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: now.Add(time.Duration(s.cfg.App.PasswordResetExpiryMinutes) * time.Minute),
		CreatedAt: now,
	}

	if err := s.userRepo.CreatePasswordResetToken(ctx, resetToken); err != nil {
		return err
	}

	resetURL := s.cfg.App.BaseURL + "/reset-password?token=" + token
	if user.Email != nil {
		go s.emailSvc.SendPasswordReset(*user.Email, user.Name, resetURL)
	}

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req *models.ResetPasswordRequest) error {
	tokenRecord, err := s.userRepo.FindPasswordResetToken(ctx, req.Token)
	if err != nil {
		return err
	}
	if tokenRecord == nil {
		return ErrInvalidToken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdatePassword(ctx, tokenRecord.UserID, string(hash)); err != nil {
		return err
	}

	if err := s.userRepo.MarkPasswordResetTokenUsed(ctx, tokenRecord.ID); err != nil {
		return err
	}

	if err := s.userRepo.RevokeAllUserRefreshTokens(ctx, tokenRecord.UserID); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) generateAuthResponse(ctx context.Context, user *models.User) (*models.AuthResponse, error) {
	identifier := ""
	if user.Email != nil {
		identifier = *user.Email
	} else if user.MobileNumber != nil {
		identifier = *user.MobileNumber
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, identifier)
	if err != nil {
		return nil, err
	}

	refreshTokenStr, err := s.jwtManager.GenerateRefreshToken(user.ID, identifier)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	refreshToken := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refreshTokenStr,
		ExpiresAt: now.Add(time.Duration(s.cfg.JWT.RefreshTokenExpiryDays) * 24 * time.Hour),
		CreatedAt: now,
	}

	if err := s.userRepo.CreateRefreshToken(ctx, refreshToken); err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		User:         user,
	}, nil
}
