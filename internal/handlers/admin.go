package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"carhistory/internal/middleware"
	"carhistory/internal/models"
	"carhistory/internal/repository"
	"carhistory/pkg/response"
)

type AdminHandler struct {
	adminRepo *repository.AdminRepository
	userRepo  *repository.UserRepository
}

func NewAdminHandler(adminRepo *repository.AdminRepository, userRepo *repository.UserRepository) *AdminHandler {
	return &AdminHandler{adminRepo: adminRepo, userRepo: userRepo}
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }
	if limit < 1 || limit > 100 { limit = 20 }

	users, total, err := h.adminRepo.ListUsers(c.Request.Context(), page, limit)
	if err != nil {
		response.InternalError(c, "failed to list users")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    users,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
			"pages": (total + limit - 1) / limit,
		},
	})
}

func (h *AdminHandler) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}
	user, err := h.userRepo.FindByID(c.Request.Context(), id)
	if err != nil || user == nil {
		response.NotFound(c, "user not found")
		return
	}
	response.OK(c, user)
}

func (h *AdminHandler) SetAdminStatus(c *gin.Context) {
	callerID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	targetID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}
	if callerID == targetID {
		response.BadRequest(c, "cannot change your own admin status")
		return
	}

	var body struct {
		IsAdmin bool `json:"is_admin"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.userRepo.SetAdminStatus(c.Request.Context(), targetID, body.IsAdmin); err != nil {
		response.InternalError(c, "failed to update admin status")
		return
	}
	response.OKMessage(c, "admin status updated")
}

func (h *AdminHandler) Analytics(c *gin.Context) {
	ctx := c.Request.Context()

	users, err := h.adminRepo.UserAnalytics(ctx)
	if err != nil {
		response.InternalError(c, "failed to fetch user analytics")
		return
	}
	vehicles, err := h.adminRepo.VehicleAnalytics(ctx)
	if err != nil {
		response.InternalError(c, "failed to fetch vehicle analytics")
		return
	}
	activity, err := h.adminRepo.ActivityAnalytics(ctx)
	if err != nil {
		response.InternalError(c, "failed to fetch activity analytics")
		return
	}

	response.OK(c, models.AdminAnalytics{
		Users:    *users,
		Vehicles: *vehicles,
		Activity: *activity,
	})
}

func (h *AdminHandler) ListAdmins(c *gin.Context) {
	admins, err := h.adminRepo.ListAdmins(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to list admins")
		return
	}
	response.OK(c, admins)
}

func (h *AdminHandler) DeactivateUser(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}
	user, err := h.userRepo.FindByID(c.Request.Context(), targetID)
	if err != nil || user == nil {
		response.NotFound(c, "user not found")
		return
	}
	if err := h.userRepo.Deactivate(c.Request.Context(), targetID); err != nil {
		response.InternalError(c, "failed to deactivate user")
		return
	}
	response.OKMessage(c, "user deactivated")
}
