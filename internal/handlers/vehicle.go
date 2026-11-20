package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"carhistory/internal/middleware"
	"carhistory/internal/models"
	"carhistory/internal/repository"
	"carhistory/internal/services"
	"carhistory/pkg/response"
)

type VehicleHandler struct {
	vehicleSvc  *services.VehicleService
	vehicleRepo *repository.VehicleRepository
	validate    *validator.Validate
}

func NewVehicleHandler(vehicleSvc *services.VehicleService, vehicleRepo *repository.VehicleRepository) *VehicleHandler {
	return &VehicleHandler{
		vehicleSvc:  vehicleSvc,
		vehicleRepo: vehicleRepo,
		validate:    validator.New(),
	}
}

func (h *VehicleHandler) GetVehicleTypes(c *gin.Context) {
	types, err := h.vehicleRepo.GetAllVehicleTypes(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to get vehicle types")
		return
	}
	response.OK(c, types)
}

func (h *VehicleHandler) SearchVehicles(c *gin.Context) {
	registration := c.Query("registration")
	email := c.Query("email")
	phone := c.Query("phone")

	if registration == "" && email == "" && phone == "" {
		response.BadRequest(c, "provide at least one search parameter: registration, email, or phone")
		return
	}

	vehicles, err := h.vehicleRepo.Search(c.Request.Context(), registration, email, phone)
	if err != nil {
		response.InternalError(c, "search failed")
		return
	}
	response.OK(c, vehicles)
}

func (h *VehicleHandler) ListVehicles(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	vehicles, err := h.vehicleSvc.ListVehicles(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "failed to list vehicles")
		return
	}
	response.OK(c, vehicles)
}

func (h *VehicleHandler) CreateVehicle(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req models.CreateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	vehicle, err := h.vehicleSvc.CreateVehicle(c.Request.Context(), userID, &req)
	if err != nil {
		response.InternalError(c, "failed to create vehicle")
		return
	}
	response.Created(c, vehicle)
}

func (h *VehicleHandler) GetVehicle(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	vehicleID, err := uuid.Parse(c.Param("vehicleId"))
	if err != nil {
		response.BadRequest(c, "invalid vehicle id")
		return
	}

	vehicle, err := h.vehicleSvc.GetVehicle(c.Request.Context(), vehicleID, userID)
	if err != nil {
		if errors.Is(err, services.ErrVehicleNotFound) {
			response.NotFound(c, "vehicle not found")
			return
		}
		response.InternalError(c, "failed to get vehicle")
		return
	}
	response.OK(c, vehicle)
}

func (h *VehicleHandler) UpdateVehicle(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	vehicleID, err := uuid.Parse(c.Param("vehicleId"))
	if err != nil {
		response.BadRequest(c, "invalid vehicle id")
		return
	}

	var req models.UpdateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	vehicle, err := h.vehicleSvc.UpdateVehicle(c.Request.Context(), vehicleID, userID, &req)
	if err != nil {
		if errors.Is(err, services.ErrVehicleNotFound) {
			response.NotFound(c, "vehicle not found")
			return
		}
		response.InternalError(c, "failed to update vehicle")
		return
	}
	response.OK(c, vehicle)
}

func (h *VehicleHandler) ArchiveVehicle(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	vehicleID, err := uuid.Parse(c.Param("vehicleId"))
	if err != nil {
		response.BadRequest(c, "invalid vehicle id")
		return
	}

	if err := h.vehicleSvc.ArchiveVehicle(c.Request.Context(), vehicleID, userID); err != nil {
		response.InternalError(c, "failed to archive vehicle")
		return
	}
	response.OKMessage(c, "vehicle archived successfully")
}

func (h *VehicleHandler) GetMembers(c *gin.Context) {
	vehicleID, err := uuid.Parse(c.Param("vehicleId"))
	if err != nil {
		response.BadRequest(c, "invalid vehicle id")
		return
	}

	members, err := h.vehicleRepo.GetMembers(c.Request.Context(), vehicleID)
	if err != nil {
		response.InternalError(c, "failed to get members")
		return
	}
	response.OK(c, members)
}

func (h *VehicleHandler) InviteMember(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	vehicleID, err := uuid.Parse(c.Param("vehicleId"))
	if err != nil {
		response.BadRequest(c, "invalid vehicle id")
		return
	}

	var req models.InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.vehicleSvc.InviteMember(c.Request.Context(), vehicleID, userID, &req); err != nil {
		if errors.Is(err, services.ErrAlreadyMember) {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c, "failed to send invitation")
		return
	}
	response.OKMessage(c, "invitation sent successfully")
}

func (h *VehicleHandler) RemoveMember(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	vehicleID, err := uuid.Parse(c.Param("vehicleId"))
	if err != nil {
		response.BadRequest(c, "invalid vehicle id")
		return
	}

	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	if err := h.vehicleSvc.RemoveMember(c.Request.Context(), vehicleID, userID, targetUserID); err != nil {
		if errors.Is(err, services.ErrCannotRemoveOwner) {
			response.Forbidden(c, err.Error())
			return
		}
		response.InternalError(c, "failed to remove member")
		return
	}
	response.OKMessage(c, "member removed successfully")
}

func (h *VehicleHandler) UpdateMemberRole(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	vehicleID, err := uuid.Parse(c.Param("vehicleId"))
	if err != nil {
		response.BadRequest(c, "invalid vehicle id")
		return
	}

	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req models.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.vehicleSvc.UpdateMemberRole(c.Request.Context(), vehicleID, userID, targetUserID, req.Role); err != nil {
		response.InternalError(c, "failed to update member role")
		return
	}
	response.OKMessage(c, "member role updated successfully")
}

func (h *VehicleHandler) GetInvitations(c *gin.Context) {
	vehicleID, err := uuid.Parse(c.Param("vehicleId"))
	if err != nil {
		response.BadRequest(c, "invalid vehicle id")
		return
	}

	invitations, err := h.vehicleRepo.FindInvitationsByVehicle(c.Request.Context(), vehicleID)
	if err != nil {
		response.InternalError(c, "failed to get invitations")
		return
	}
	response.OK(c, invitations)
}

func (h *VehicleHandler) GetPendingInvitations(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	emailRaw, _ := c.Get(middleware.ContextUserEmail)
	email, _ := emailRaw.(string)
	_ = userID

	invitations, err := h.vehicleRepo.FindPendingInvitationsByEmail(c.Request.Context(), email)
	if err != nil {
		response.InternalError(c, "failed to get invitations")
		return
	}
	response.OK(c, invitations)
}

func (h *VehicleHandler) AcceptInvitation(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	token := c.Param("token")
	if err := h.vehicleSvc.AcceptInvitation(c.Request.Context(), token, userID); err != nil {
		if errors.Is(err, services.ErrInvitationInvalid) || errors.Is(err, services.ErrAlreadyMember) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, "failed to accept invitation")
		return
	}
	response.OKMessage(c, "invitation accepted successfully")
}

func (h *VehicleHandler) RejectInvitation(c *gin.Context) {
	token := c.Param("token")
	if err := h.vehicleSvc.RejectInvitation(c.Request.Context(), token); err != nil {
		response.BadRequest(c, "invalid or expired invitation")
		return
	}
	response.OKMessage(c, "invitation rejected")
}
