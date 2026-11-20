package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"carhistory/internal/middleware"
	"carhistory/internal/models"
	"carhistory/internal/repository"
	"carhistory/pkg/response"
)

type FuelHandler struct {
	fuelRepo *repository.FuelRepository
	validate *validator.Validate
}

func NewFuelHandler(fuelRepo *repository.FuelRepository) *FuelHandler {
	return &FuelHandler{fuelRepo: fuelRepo, validate: validator.New()}
}

func (h *FuelHandler) Create(c *gin.Context) {
	vehicleID, err := uuid.Parse(c.Param("vehicleId"))
	if err != nil {
		response.BadRequest(c, "invalid vehicle id")
		return
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req models.CreateFuelLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	now := time.Now()
	log := &models.FuelLog{
		ID:              uuid.New(),
		VehicleID:       vehicleID,
		LoggedBy:        userID,
		FuelType:        req.FuelType,
		Quantity:        req.Quantity,
		PricePerUnit:    req.PricePerUnit,
		TotalCost:       req.TotalCost,
		OdometerReading: req.OdometerReading,
		FuelStation:     req.FuelStation,
		Date:            req.Date,
		Notes:           req.Notes,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := h.fuelRepo.Create(c.Request.Context(), log); err != nil {
		response.InternalError(c, "failed to create fuel log")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": log})
}

func (h *FuelHandler) List(c *gin.Context) {
	vehicleID, err := uuid.Parse(c.Param("vehicleId"))
	if err != nil {
		response.BadRequest(c, "invalid vehicle id")
		return
	}
	logs, err := h.fuelRepo.List(c.Request.Context(), vehicleID)
	if err != nil {
		response.InternalError(c, "failed to list fuel logs")
		return
	}
	response.OK(c, logs)
}

func (h *FuelHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	log, err := h.fuelRepo.FindByID(c.Request.Context(), id)
	if err != nil || log == nil {
		response.NotFound(c, "fuel log not found")
		return
	}
	response.OK(c, log)
}

func (h *FuelHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	existing, err := h.fuelRepo.FindByID(c.Request.Context(), id)
	if err != nil || existing == nil {
		response.NotFound(c, "fuel log not found")
		return
	}

	var req models.UpdateFuelLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.FuelType != nil { existing.FuelType = *req.FuelType }
	if req.Quantity != nil { existing.Quantity = req.Quantity }
	if req.PricePerUnit != nil { existing.PricePerUnit = req.PricePerUnit }
	if req.TotalCost != nil { existing.TotalCost = *req.TotalCost }
	if req.OdometerReading != nil { existing.OdometerReading = req.OdometerReading }
	if req.FuelStation != nil { existing.FuelStation = req.FuelStation }
	if req.Date != nil { existing.Date = *req.Date }
	if req.Notes != nil { existing.Notes = req.Notes }
	existing.UpdatedAt = time.Now()

	if err := h.fuelRepo.Update(c.Request.Context(), existing); err != nil {
		response.InternalError(c, "failed to update fuel log")
		return
	}
	response.OK(c, existing)
}

func (h *FuelHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	if err := h.fuelRepo.Delete(c.Request.Context(), id); err != nil {
		response.InternalError(c, "failed to delete fuel log")
		return
	}
	response.OKMessage(c, "fuel log deleted")
}

func (h *FuelHandler) Summary(c *gin.Context) {
	vehicleID, err := uuid.Parse(c.Param("vehicleId"))
	if err != nil {
		response.BadRequest(c, "invalid vehicle id")
		return
	}
	summary, err := h.fuelRepo.Summary(c.Request.Context(), vehicleID)
	if err != nil {
		response.InternalError(c, "failed to calculate fuel summary")
		return
	}
	response.OK(c, summary)
}
