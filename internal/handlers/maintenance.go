package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"carhistory/internal/middleware"
	"carhistory/internal/models"
	"carhistory/internal/repository"
	"carhistory/pkg/response"
	"carhistory/pkg/utils"
)

type MaintenanceHandler struct {
	maintenanceRepo *repository.MaintenanceRepository
	activityRepo    *repository.ActivityRepository
	validate        *validator.Validate
}

func NewMaintenanceHandler(maintenanceRepo *repository.MaintenanceRepository, activityRepo *repository.ActivityRepository) *MaintenanceHandler {
	return &MaintenanceHandler{
		maintenanceRepo: maintenanceRepo,
		activityRepo:    activityRepo,
		validate:        validator.New(),
	}
}

func (h *MaintenanceHandler) GetTypes(c *gin.Context) {
	types, err := h.maintenanceRepo.GetTypes(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to get maintenance types")
		return
	}
	response.OK(c, types)
}

func (h *MaintenanceHandler) List(c *gin.Context) {
	vehicleID, err := uuid.Parse(c.Param("vehicleId"))
	if err != nil {
		response.BadRequest(c, "invalid vehicle id")
		return
	}

	page, limit, offset := utils.GetPagination(c)
	records, total, err := h.maintenanceRepo.FindByVehicle(c.Request.Context(), vehicleID, limit, offset)
	if err != nil {
		response.InternalError(c, "failed to list maintenance records")
		return
	}
	response.Paginated(c, records, total, page, limit)
}

func (h *MaintenanceHandler) Create(c *gin.Context) {
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

	var req models.CreateMaintenanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	now := time.Now()
	rec := &models.MaintenanceRecord{
		ID:                    uuid.New(),
		VehicleID:             vehicleID,
		MaintenanceTypeID:     req.MaintenanceTypeID,
		Title:                 req.Title,
		DatePerformed:         req.DatePerformed,
		OdometerAtMaintenance: req.OdometerAtMaintenance,
		NextDueDate:           req.NextDueDate,
		NextDueOdometer:       req.NextDueOdometer,
		Cost:                  req.Cost,
		ServiceProvider:       req.ServiceProvider,
		Description:           req.Description,
		Remarks:               req.Remarks,
		CreatedBy:             userID,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if err := h.maintenanceRepo.Create(c.Request.Context(), rec); err != nil {
		response.InternalError(c, "failed to create maintenance record")
		return
	}

	desc := "Added maintenance record: " + rec.Title
	h.logActivity(c, &vehicleID, userID, "added_maintenance", "maintenance_record", &rec.ID, desc)

	created, _ := h.maintenanceRepo.FindByID(c.Request.Context(), rec.ID, vehicleID)
	response.Created(c, created)
}

func (h *MaintenanceHandler) Get(c *gin.Context) {
	vehicleID, err := uuid.Parse(c.Param("vehicleId"))
	if err != nil {
		response.BadRequest(c, "invalid vehicle id")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid maintenance record id")
		return
	}

	rec, err := h.maintenanceRepo.FindByID(c.Request.Context(), id, vehicleID)
	if err != nil {
		response.InternalError(c, "failed to get maintenance record")
		return
	}
	if rec == nil {
		response.NotFound(c, "maintenance record not found")
		return
	}
	response.OK(c, rec)
}

func (h *MaintenanceHandler) Update(c *gin.Context) {
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

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid maintenance record id")
		return
	}

	rec, err := h.maintenanceRepo.FindByID(c.Request.Context(), id, vehicleID)
	if err != nil || rec == nil {
		response.NotFound(c, "maintenance record not found")
		return
	}

	var req models.UpdateMaintenanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.MaintenanceTypeID != nil {
		rec.MaintenanceTypeID = *req.MaintenanceTypeID
	}
	if req.Title != nil {
		rec.Title = *req.Title
	}
	if req.DatePerformed != nil {
		rec.DatePerformed = *req.DatePerformed
	}
	if req.OdometerAtMaintenance != nil {
		rec.OdometerAtMaintenance = req.OdometerAtMaintenance
	}
	if req.NextDueDate != nil {
		rec.NextDueDate = req.NextDueDate
	}
	if req.NextDueOdometer != nil {
		rec.NextDueOdometer = req.NextDueOdometer
	}
	if req.Cost != nil {
		rec.Cost = req.Cost
	}
	if req.ServiceProvider != nil {
		rec.ServiceProvider = req.ServiceProvider
	}
	if req.Description != nil {
		rec.Description = req.Description
	}
	if req.Remarks != nil {
		rec.Remarks = req.Remarks
	}
	rec.UpdatedAt = time.Now()

	if err := h.maintenanceRepo.Update(c.Request.Context(), rec); err != nil {
		response.InternalError(c, "failed to update maintenance record")
		return
	}

	desc := "Updated maintenance record: " + rec.Title
	h.logActivity(c, &vehicleID, userID, "updated_maintenance", "maintenance_record", &rec.ID, desc)

	response.OK(c, rec)
}

func (h *MaintenanceHandler) Delete(c *gin.Context) {
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

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid maintenance record id")
		return
	}

	rec, err := h.maintenanceRepo.FindByID(c.Request.Context(), id, vehicleID)
	if err != nil || rec == nil {
		response.NotFound(c, "maintenance record not found")
		return
	}

	if err := h.maintenanceRepo.Delete(c.Request.Context(), id, vehicleID); err != nil {
		response.InternalError(c, "failed to delete maintenance record")
		return
	}

	desc := "Deleted maintenance record: " + rec.Title
	h.logActivity(c, &vehicleID, userID, "deleted_maintenance", "maintenance_record", &id, desc)

	response.OKMessage(c, "maintenance record deleted")
}

func (h *MaintenanceHandler) logActivity(c *gin.Context, vehicleID *uuid.UUID, userID uuid.UUID, action, entity string, entityID *uuid.UUID, description string) {
	log := &models.ActivityLog{
		ID:          uuid.New(),
		VehicleID:   vehicleID,
		UserID:      userID,
		Action:      action,
		Entity:      &entity,
		EntityID:    entityID,
		Description: &description,
		CreatedAt:   time.Now(),
	}
	_ = h.activityRepo.Create(c.Request.Context(), log)
}
