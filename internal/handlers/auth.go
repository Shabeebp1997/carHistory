package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"carhistory/internal/models"
	"carhistory/internal/repository"
	"carhistory/internal/services"
	jwtpkg "carhistory/pkg/jwt"
	"carhistory/pkg/response"
)

type AuthHandler struct {
	authSvc    *services.AuthService
	userRepo   *repository.UserRepository
	jwtManager *jwtpkg.Manager
	validate   *validator.Validate
}

func NewAuthHandler(authSvc *services.AuthService, userRepo *repository.UserRepository, jwtManager *jwtpkg.Manager) *AuthHandler {
	return &AuthHandler{
		authSvc:    authSvc,
		userRepo:   userRepo,
		jwtManager: jwtManager,
		validate:   validator.New(),
	}
}

func (h *AuthHandler) callerIsAdmin(c *gin.Context) bool {
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return false
	}
	claims, err := h.jwtManager.ValidateToken(strings.TrimPrefix(auth, "Bearer "))
	if err != nil {
		return false
	}
	user, err := h.userRepo.FindByID(c.Request.Context(), claims.UserID)
	return err == nil && user != nil && user.IsAdmin
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.IsAdmin && !h.callerIsAdmin(c) {
		response.Forbidden(c, "only an existing admin can create an admin account")
		return
	}

	res, err := h.authSvc.Register(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEmailTaken),
			errors.Is(err, services.ErrPhoneTaken):
			response.Conflict(c, err.Error())
		case errors.Is(err, services.ErrContactRequired):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, "registration failed")
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": res})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	res, err := h.authSvc.Login(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			response.Unauthorized(c, "invalid credentials")
		case errors.Is(err, services.ErrContactRequired):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, "login failed")
		}
		return
	}

	response.OK(c, res)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.authSvc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		response.InternalError(c, "logout failed")
		return
	}

	response.OKMessage(c, "logged out successfully")
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	res, err := h.authSvc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, "invalid or expired refresh token")
		return
	}

	response.OK(c, res)
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	_ = h.authSvc.ForgotPassword(c.Request.Context(), req.Email)
	response.OKMessage(c, "if the email exists, a reset link has been sent")
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.authSvc.ResetPassword(c.Request.Context(), &req); err != nil {
		if errors.Is(err, services.ErrInvalidToken) {
			response.BadRequest(c, "invalid or expired reset token")
			return
		}
		response.InternalError(c, "password reset failed")
		return
	}

	response.OKMessage(c, "password reset successfully")
}

