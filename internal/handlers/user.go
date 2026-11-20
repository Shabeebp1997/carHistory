package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"carhistory/internal/config"
	"carhistory/internal/middleware"
	"carhistory/internal/models"
	"carhistory/internal/repository"
	"carhistory/internal/services"
	"carhistory/pkg/response"
)

type UserHandler struct {
	userSvc  *services.UserService
	userRepo *repository.UserRepository
	cfg      *config.Config
	validate *validator.Validate
}

func NewUserHandler(userSvc *services.UserService, userRepo *repository.UserRepository, cfg *config.Config) *UserHandler {
	return &UserHandler{
		userSvc:  userSvc,
		userRepo: userRepo,
		cfg:      cfg,
		validate: validator.New(),
	}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	user, err := h.userSvc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "failed to get profile")
		return
	}

	response.OK(c, user)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.userSvc.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		response.InternalError(c, "failed to update profile")
		return
	}

	response.OK(c, user)
}

func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}
	defer file.Close()

	mimeType := header.Header.Get("Content-Type")
	validImages := map[string]string{
		"image/jpeg": "jpg", "image/png": "png",
		"image/gif": "gif", "image/webp": "webp",
	}
	ext, allowed := validImages[mimeType]
	if !allowed {
		response.BadRequest(c, "only jpeg, png, gif, webp images are allowed")
		return
	}

	dir := filepath.Join(h.cfg.Storage.LocalPath, "avatars")
	if err := os.MkdirAll(dir, 0755); err != nil {
		response.InternalError(c, "failed to create avatar directory")
		return
	}

	fileName := fmt.Sprintf("%s.%s", userID.String(), ext)
	savePath := filepath.Join(dir, fileName)
	out, err := os.Create(savePath)
	if err != nil {
		response.InternalError(c, "failed to save avatar")
		return
	}
	defer out.Close()
	if _, err = io.Copy(out, file); err != nil {
		response.InternalError(c, "failed to write avatar")
		return
	}

	avatarURL := strings.TrimRight(h.cfg.Storage.BaseURL, "/") + "/avatars/" + fileName
	if err := h.userRepo.UpdateProfilePicture(c.Request.Context(), userID, avatarURL); err != nil {
		response.InternalError(c, "failed to update profile picture")
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "profile_picture": avatarURL})
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.userSvc.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		if err == services.ErrInvalidCredentials {
			response.BadRequest(c, "current password is incorrect")
			return
		}
		response.InternalError(c, "failed to change password")
		return
	}

	response.OKMessage(c, "password changed successfully")
}
