package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"carhistory/internal/config"
	"carhistory/internal/middleware"
	"carhistory/internal/models"
	"carhistory/internal/repository"
	"carhistory/pkg/response"
)

var allowedMIME = map[string]string{
	"image/jpeg":      "jpg",
	"image/png":       "png",
	"image/gif":       "gif",
	"image/webp":      "webp",
	"video/mp4":       "mp4",
	"video/quicktime": "mov",
	"video/x-msvideo": "avi",
}

type MediaHandler struct {
	mediaRepo *repository.MediaRepository
	cfg       *config.Config
}

func NewMediaHandler(mediaRepo *repository.MediaRepository, cfg *config.Config) *MediaHandler {
	return &MediaHandler{mediaRepo: mediaRepo, cfg: cfg}
}

func (h *MediaHandler) Upload(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	entityType := c.PostForm("entity_type")
	entityIDStr := c.PostForm("entity_id")

	if !models.ValidEntityTypes[entityType] {
		response.BadRequest(c, "invalid entity_type")
		return
	}
	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		response.BadRequest(c, "invalid entity_id")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}
	defer file.Close()

	mimeType := header.Header.Get("Content-Type")
	ext, allowed := allowedMIME[mimeType]
	if !allowed {
		response.BadRequest(c, "unsupported file type; allowed: jpeg, png, gif, webp, mp4, mov, avi")
		return
	}

	mediaType := "image"
	if strings.HasPrefix(mimeType, "video/") {
		mediaType = "video"
	}

	dir := filepath.Join(h.cfg.Storage.LocalPath, entityType)
	if err := os.MkdirAll(dir, 0755); err != nil {
		response.InternalError(c, "failed to create storage directory")
		return
	}

	newName := fmt.Sprintf("%s_%d.%s", uuid.New().String(), time.Now().UnixMilli(), ext)
	savePath := filepath.Join(dir, newName)

	out, err := os.Create(savePath)
	if err != nil {
		response.InternalError(c, "failed to save file")
		return
	}
	defer out.Close()

	size, err := io.Copy(out, file)
	if err != nil {
		response.InternalError(c, "failed to write file")
		return
	}

	mime := mimeType
	media := &models.Media{
		ID:         uuid.New(),
		EntityType: entityType,
		EntityID:   entityID,
		FilePath:   savePath,
		FileName:   header.Filename,
		MediaType:  mediaType,
		MimeType:   &mime,
		FileSize:   size,
		UploadedBy: &userID,
		CreatedAt:  time.Now(),
	}

	if err := h.mediaRepo.Create(c.Request.Context(), media); err != nil {
		response.InternalError(c, "failed to save media record")
		return
	}

	media.FileURL = h.cfg.Storage.BaseURL + "/" + entityType + "/" + newName
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": media})
}

func (h *MediaHandler) ListByEntity(c *gin.Context) {
	entityType := c.Param("entityType")
	entityIDStr := c.Param("entityId")

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		response.BadRequest(c, "invalid entity_id")
		return
	}

	items, err := h.mediaRepo.FindByEntity(c.Request.Context(), entityType, entityID)
	if err != nil {
		response.InternalError(c, "failed to fetch media")
		return
	}

	for i := range items {
		parts := strings.Split(filepath.ToSlash(items[i].FilePath), "/")
		items[i].FileURL = h.cfg.Storage.BaseURL + "/" + strings.Join(parts[len(parts)-2:], "/")
	}
	response.OK(c, items)
}

func (h *MediaHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	m, err := h.mediaRepo.FindByID(c.Request.Context(), id)
	if err != nil || m == nil {
		response.NotFound(c, "media not found")
		return
	}

	_ = os.Remove(m.FilePath)

	if err := h.mediaRepo.Delete(c.Request.Context(), id); err != nil {
		response.InternalError(c, "failed to delete media")
		return
	}
	response.OKMessage(c, "media deleted")
}

