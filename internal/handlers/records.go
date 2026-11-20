package handlers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"carhistory/internal/config"
	"carhistory/internal/middleware"
	"carhistory/internal/models"
	"carhistory/internal/repository"
	"carhistory/pkg/response"
	"carhistory/pkg/utils"
)

// ============================================================
// SERVICE RECORD HANDLER
// ============================================================

type ServiceRecordHandler struct {
	serviceRepo  *repository.ServiceRepository
	activityRepo *repository.ActivityRepository
	validate     *validator.Validate
}

func NewServiceRecordHandler(serviceRepo *repository.ServiceRepository, activityRepo *repository.ActivityRepository) *ServiceRecordHandler {
	return &ServiceRecordHandler{serviceRepo: serviceRepo, activityRepo: activityRepo, validate: validator.New()}
}

func (h *ServiceRecordHandler) List(c *gin.Context) {
	vehicleID, err := uuid.Parse(c.Param("vehicleId"))
	if err != nil {
		response.BadRequest(c, "invalid vehicle id")
		return
	}
	page, limit, offset := utils.GetPagination(c)
	records, total, err := h.serviceRepo.FindByVehicle(c.Request.Context(), vehicleID, limit, offset)
	if err != nil {
		response.InternalError(c, "failed to list service records")
		return
	}
	response.Paginated(c, records, total, page, limit)
}

func (h *ServiceRecordHandler) Create(c *gin.Context) {
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

	var req models.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	now := time.Now()
	total := req.LabourCost + req.PartsCost + req.OtherCost
	for _, item := range req.Items {
		total += item.Cost
	}

	rec := &models.ServiceRecord{
		ID:              uuid.New(),
		VehicleID:       vehicleID,
		ServiceDate:     req.ServiceDate,
		Odometer:        req.Odometer,
		ServiceType:     req.ServiceType,
		ServiceCentre:   req.ServiceCentre,
		Technician:      req.Technician,
		LabourCost:      req.LabourCost,
		PartsCost:       req.PartsCost,
		OtherCost:       req.OtherCost,
		TotalCost:       total,
		NextServiceDate: req.NextServiceDate,
		NextServiceKM:   req.NextServiceKM,
		Remarks:         req.Remarks,
		CreatedBy:       userID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := h.serviceRepo.Create(c.Request.Context(), rec); err != nil {
		response.InternalError(c, "failed to create service record")
		return
	}

	for _, item := range req.Items {
		si := &models.ServiceItem{
			ID:              uuid.New(),
			ServiceRecordID: rec.ID,
			Name:            item.Name,
			Description:     item.Description,
			Cost:            item.Cost,
			CreatedAt:       now,
		}
		_ = h.serviceRepo.CreateItem(c.Request.Context(), si)
	}

	desc := "Added service record"
	if req.ServiceCentre != nil {
		desc += " at " + *req.ServiceCentre
	}
	h.logActivity(c, &vehicleID, userID, "added_service", "service_record", &rec.ID, desc)

	created, _ := h.serviceRepo.FindByID(c.Request.Context(), rec.ID, vehicleID)
	response.Created(c, created)
}

func (h *ServiceRecordHandler) Get(c *gin.Context) {
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid service record id")
		return
	}
	rec, err := h.serviceRepo.FindByID(c.Request.Context(), id, vehicleID)
	if err != nil || rec == nil {
		response.NotFound(c, "service record not found")
		return
	}
	response.OK(c, rec)
}

func (h *ServiceRecordHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid service record id")
		return
	}

	rec, err := h.serviceRepo.FindByID(c.Request.Context(), id, vehicleID)
	if err != nil || rec == nil {
		response.NotFound(c, "service record not found")
		return
	}

	var req models.UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.ServiceDate != nil {
		rec.ServiceDate = *req.ServiceDate
	}
	if req.Odometer != nil {
		rec.Odometer = req.Odometer
	}
	if req.ServiceType != nil {
		rec.ServiceType = req.ServiceType
	}
	if req.ServiceCentre != nil {
		rec.ServiceCentre = req.ServiceCentre
	}
	if req.Technician != nil {
		rec.Technician = req.Technician
	}
	if req.LabourCost != nil {
		rec.LabourCost = *req.LabourCost
	}
	if req.PartsCost != nil {
		rec.PartsCost = *req.PartsCost
	}
	if req.OtherCost != nil {
		rec.OtherCost = *req.OtherCost
	}
	rec.TotalCost = rec.LabourCost + rec.PartsCost + rec.OtherCost
	if req.NextServiceDate != nil {
		rec.NextServiceDate = req.NextServiceDate
	}
	if req.NextServiceKM != nil {
		rec.NextServiceKM = req.NextServiceKM
	}
	if req.Remarks != nil {
		rec.Remarks = req.Remarks
	}
	rec.UpdatedAt = time.Now()

	if err := h.serviceRepo.Update(c.Request.Context(), rec); err != nil {
		response.InternalError(c, "failed to update service record")
		return
	}

	if req.Items != nil {
		_ = h.serviceRepo.DeleteItems(c.Request.Context(), rec.ID)
		for _, item := range req.Items {
			si := &models.ServiceItem{
				ID:              uuid.New(),
				ServiceRecordID: rec.ID,
				Name:            item.Name,
				Description:     item.Description,
				Cost:            item.Cost,
				CreatedAt:       time.Now(),
			}
			_ = h.serviceRepo.CreateItem(c.Request.Context(), si)
		}
	}

	h.logActivity(c, &vehicleID, userID, "updated_service", "service_record", &rec.ID, "Updated service record")

	updated, _ := h.serviceRepo.FindByID(c.Request.Context(), rec.ID, vehicleID)
	response.OK(c, updated)
}

func (h *ServiceRecordHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid service record id")
		return
	}
	if err := h.serviceRepo.Delete(c.Request.Context(), id, vehicleID); err != nil {
		response.InternalError(c, "failed to delete service record")
		return
	}
	h.logActivity(c, &vehicleID, userID, "deleted_service", "service_record", &id, "Deleted service record")
	response.OKMessage(c, "service record deleted")
}

func (h *ServiceRecordHandler) logActivity(c *gin.Context, vehicleID *uuid.UUID, userID uuid.UUID, action, entity string, entityID *uuid.UUID, description string) {
	log := &models.ActivityLog{
		ID: uuid.New(), VehicleID: vehicleID, UserID: userID,
		Action: action, Entity: &entity, EntityID: entityID, Description: &description, CreatedAt: time.Now(),
	}
	_ = h.activityRepo.Create(c.Request.Context(), log)
}

// ============================================================
// INSURANCE HANDLER
// ============================================================

type InsuranceHandler struct {
	insuranceRepo *repository.InsuranceRepository
	activityRepo  *repository.ActivityRepository
	validate      *validator.Validate
}

func NewInsuranceHandler(insuranceRepo *repository.InsuranceRepository, activityRepo *repository.ActivityRepository) *InsuranceHandler {
	return &InsuranceHandler{insuranceRepo: insuranceRepo, activityRepo: activityRepo, validate: validator.New()}
}

func (h *InsuranceHandler) List(c *gin.Context) {
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	records, err := h.insuranceRepo.FindByVehicle(c.Request.Context(), vehicleID)
	if err != nil {
		response.InternalError(c, "failed to list insurance records")
		return
	}
	for i := range records {
		records[i].Status = insuranceStatus(records[i].ExpiryDate)
	}
	response.OK(c, records)
}

func (h *InsuranceHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))

	var req models.CreateInsuranceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	now := time.Now()
	rec := &models.InsuranceRecord{
		ID:               uuid.New(),
		VehicleID:        vehicleID,
		InsuranceCompany: req.InsuranceCompany,
		PolicyNumber:     req.PolicyNumber,
		PolicyType:       req.PolicyType,
		StartDate:        req.StartDate,
		ExpiryDate:       req.ExpiryDate,
		Premium:          req.Premium,
		CoverageAmount:   req.CoverageAmount,
		Remarks:          req.Remarks,
		CreatedBy:        userID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := h.insuranceRepo.Create(c.Request.Context(), rec); err != nil {
		response.InternalError(c, "failed to create insurance record")
		return
	}

	desc := "Added insurance record: " + rec.InsuranceCompany
	h.logActivity(c, &vehicleID, userID, "added_insurance", "insurance_record", &rec.ID, desc)

	rec.Status = insuranceStatus(rec.ExpiryDate)
	response.Created(c, rec)
}

func (h *InsuranceHandler) Get(c *gin.Context) {
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	rec, err := h.insuranceRepo.FindByID(c.Request.Context(), id, vehicleID)
	if err != nil || rec == nil {
		response.NotFound(c, "insurance record not found")
		return
	}
	rec.Status = insuranceStatus(rec.ExpiryDate)
	response.OK(c, rec)
}

func (h *InsuranceHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, _ := uuid.Parse(c.Param("id"))

	rec, err := h.insuranceRepo.FindByID(c.Request.Context(), id, vehicleID)
	if err != nil || rec == nil {
		response.NotFound(c, "insurance record not found")
		return
	}

	var req models.UpdateInsuranceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.InsuranceCompany != nil {
		rec.InsuranceCompany = *req.InsuranceCompany
	}
	if req.PolicyNumber != nil {
		rec.PolicyNumber = req.PolicyNumber
	}
	if req.PolicyType != nil {
		rec.PolicyType = req.PolicyType
	}
	if req.StartDate != nil {
		rec.StartDate = *req.StartDate
	}
	if req.ExpiryDate != nil {
		rec.ExpiryDate = *req.ExpiryDate
	}
	if req.Premium != nil {
		rec.Premium = req.Premium
	}
	if req.CoverageAmount != nil {
		rec.CoverageAmount = req.CoverageAmount
	}
	if req.Remarks != nil {
		rec.Remarks = req.Remarks
	}
	rec.UpdatedAt = time.Now()

	if err := h.insuranceRepo.Update(c.Request.Context(), rec); err != nil {
		response.InternalError(c, "failed to update insurance record")
		return
	}
	h.logActivity(c, &vehicleID, userID, "updated_insurance", "insurance_record", &rec.ID, "Updated insurance record")
	rec.Status = insuranceStatus(rec.ExpiryDate)
	response.OK(c, rec)
}

func (h *InsuranceHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.insuranceRepo.Delete(c.Request.Context(), id, vehicleID); err != nil {
		response.InternalError(c, "failed to delete insurance record")
		return
	}
	h.logActivity(c, &vehicleID, userID, "deleted_insurance", "insurance_record", &id, "Deleted insurance record")
	response.OKMessage(c, "insurance record deleted")
}

func (h *InsuranceHandler) logActivity(c *gin.Context, vehicleID *uuid.UUID, userID uuid.UUID, action, entity string, entityID *uuid.UUID, description string) {
	log := &models.ActivityLog{
		ID: uuid.New(), VehicleID: vehicleID, UserID: userID,
		Action: action, Entity: &entity, EntityID: entityID, Description: &description, CreatedAt: time.Now(),
	}
	_ = h.activityRepo.Create(c.Request.Context(), log)
}

func insuranceStatus(expiry time.Time) string {
	now := time.Now()
	if expiry.Before(now) {
		return "expired"
	}
	if expiry.Before(now.Add(30 * 24 * time.Hour)) {
		return "expiring_soon"
	}
	return "valid"
}

// ============================================================
// PUC HANDLER
// ============================================================

type PUCHandler struct {
	pucRepo      *repository.PUCRepository
	activityRepo *repository.ActivityRepository
	validate     *validator.Validate
}

func NewPUCHandler(pucRepo *repository.PUCRepository, activityRepo *repository.ActivityRepository) *PUCHandler {
	return &PUCHandler{pucRepo: pucRepo, activityRepo: activityRepo, validate: validator.New()}
}

func (h *PUCHandler) List(c *gin.Context) {
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	records, err := h.pucRepo.FindByVehicle(c.Request.Context(), vehicleID)
	if err != nil {
		response.InternalError(c, "failed to list PUC records")
		return
	}
	for i := range records {
		records[i].Status = insuranceStatus(records[i].ExpiryDate)
	}
	response.OK(c, records)
}

func (h *PUCHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))

	var req models.CreatePUCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	now := time.Now()
	rec := &models.PUCRecord{
		ID:                uuid.New(),
		VehicleID:         vehicleID,
		CertificateNumber: req.CertificateNumber,
		TestDate:          req.TestDate,
		ExpiryDate:        req.ExpiryDate,
		TestingCentre:     req.TestingCentre,
		EmissionInfo:      req.EmissionInfo,
		Remarks:           req.Remarks,
		CreatedBy:         userID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := h.pucRepo.Create(c.Request.Context(), rec); err != nil {
		response.InternalError(c, "failed to create PUC record")
		return
	}

	h.logActivity(c, &vehicleID, userID, "added_puc", "puc_record", &rec.ID, "Added PUC record")
	rec.Status = insuranceStatus(rec.ExpiryDate)
	response.Created(c, rec)
}

func (h *PUCHandler) Get(c *gin.Context) {
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, _ := uuid.Parse(c.Param("id"))
	rec, err := h.pucRepo.FindByID(c.Request.Context(), id, vehicleID)
	if err != nil || rec == nil {
		response.NotFound(c, "PUC record not found")
		return
	}
	rec.Status = insuranceStatus(rec.ExpiryDate)
	response.OK(c, rec)
}

func (h *PUCHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, _ := uuid.Parse(c.Param("id"))

	rec, err := h.pucRepo.FindByID(c.Request.Context(), id, vehicleID)
	if err != nil || rec == nil {
		response.NotFound(c, "PUC record not found")
		return
	}

	var req models.UpdatePUCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.CertificateNumber != nil {
		rec.CertificateNumber = req.CertificateNumber
	}
	if req.TestDate != nil {
		rec.TestDate = *req.TestDate
	}
	if req.ExpiryDate != nil {
		rec.ExpiryDate = *req.ExpiryDate
	}
	if req.TestingCentre != nil {
		rec.TestingCentre = req.TestingCentre
	}
	if req.EmissionInfo != nil {
		rec.EmissionInfo = req.EmissionInfo
	}
	if req.Remarks != nil {
		rec.Remarks = req.Remarks
	}
	rec.UpdatedAt = time.Now()

	if err := h.pucRepo.Update(c.Request.Context(), rec); err != nil {
		response.InternalError(c, "failed to update PUC record")
		return
	}
	h.logActivity(c, &vehicleID, userID, "updated_puc", "puc_record", &rec.ID, "Updated PUC record")
	rec.Status = insuranceStatus(rec.ExpiryDate)
	response.OK(c, rec)
}

func (h *PUCHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.pucRepo.Delete(c.Request.Context(), id, vehicleID); err != nil {
		response.InternalError(c, "failed to delete PUC record")
		return
	}
	h.logActivity(c, &vehicleID, userID, "deleted_puc", "puc_record", &id, "Deleted PUC record")
	response.OKMessage(c, "PUC record deleted")
}

func (h *PUCHandler) logActivity(c *gin.Context, vehicleID *uuid.UUID, userID uuid.UUID, action, entity string, entityID *uuid.UUID, description string) {
	log := &models.ActivityLog{
		ID: uuid.New(), VehicleID: vehicleID, UserID: userID,
		Action: action, Entity: &entity, EntityID: entityID, Description: &description, CreatedAt: time.Now(),
	}
	_ = h.activityRepo.Create(c.Request.Context(), log)
}

// ============================================================
// TYRE HANDLER
// ============================================================

type TyreHandler struct {
	tyreRepo     *repository.TyreRepository
	activityRepo *repository.ActivityRepository
	validate     *validator.Validate
}

func NewTyreHandler(tyreRepo *repository.TyreRepository, activityRepo *repository.ActivityRepository) *TyreHandler {
	return &TyreHandler{tyreRepo: tyreRepo, activityRepo: activityRepo, validate: validator.New()}
}

func (h *TyreHandler) List(c *gin.Context) {
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	records, err := h.tyreRepo.FindByVehicle(c.Request.Context(), vehicleID)
	if err != nil {
		response.InternalError(c, "failed to list tyre records")
		return
	}
	response.OK(c, records)
}

func (h *TyreHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))

	var req models.CreateTyreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	now := time.Now()
	rec := &models.TyreRecord{
		ID:               uuid.New(),
		VehicleID:        vehicleID,
		Brand:            req.Brand,
		Model:            req.Model,
		Size:             req.Size,
		Position:         req.Position,
		PurchaseDate:     req.PurchaseDate,
		InstallationDate: req.InstallationDate,
		InstallationKM:   req.InstallationKM,
		RemovalDate:      req.RemovalDate,
		RemovalKM:        req.RemovalKM,
		Cost:             req.Cost,
		WarrantyExpiry:   req.WarrantyExpiry,
		Remarks:          req.Remarks,
		CreatedBy:        userID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := h.tyreRepo.Create(c.Request.Context(), rec); err != nil {
		response.InternalError(c, "failed to create tyre record")
		return
	}

	h.logActivity(c, &vehicleID, userID, "added_tyre", "tyre_record", &rec.ID, "Added tyre record")
	response.Created(c, rec)
}

func (h *TyreHandler) Get(c *gin.Context) {
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, _ := uuid.Parse(c.Param("id"))
	rec, err := h.tyreRepo.FindByID(c.Request.Context(), id, vehicleID)
	if err != nil || rec == nil {
		response.NotFound(c, "tyre record not found")
		return
	}
	response.OK(c, rec)
}

func (h *TyreHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, _ := uuid.Parse(c.Param("id"))

	rec, err := h.tyreRepo.FindByID(c.Request.Context(), id, vehicleID)
	if err != nil || rec == nil {
		response.NotFound(c, "tyre record not found")
		return
	}

	var req models.UpdateTyreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Brand != nil {
		rec.Brand = req.Brand
	}
	if req.Model != nil {
		rec.Model = req.Model
	}
	if req.Size != nil {
		rec.Size = req.Size
	}
	if req.Position != nil {
		rec.Position = req.Position
	}
	if req.PurchaseDate != nil {
		rec.PurchaseDate = req.PurchaseDate
	}
	if req.InstallationDate != nil {
		rec.InstallationDate = req.InstallationDate
	}
	if req.InstallationKM != nil {
		rec.InstallationKM = req.InstallationKM
	}
	if req.RemovalDate != nil {
		rec.RemovalDate = req.RemovalDate
	}
	if req.RemovalKM != nil {
		rec.RemovalKM = req.RemovalKM
	}
	if req.Cost != nil {
		rec.Cost = req.Cost
	}
	if req.WarrantyExpiry != nil {
		rec.WarrantyExpiry = req.WarrantyExpiry
	}
	if req.Remarks != nil {
		rec.Remarks = req.Remarks
	}
	rec.UpdatedAt = time.Now()

	if err := h.tyreRepo.Update(c.Request.Context(), rec); err != nil {
		response.InternalError(c, "failed to update tyre record")
		return
	}
	h.logActivity(c, &vehicleID, userID, "updated_tyre", "tyre_record", &rec.ID, "Updated tyre record")
	response.OK(c, rec)
}

func (h *TyreHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.tyreRepo.Delete(c.Request.Context(), id, vehicleID); err != nil {
		response.InternalError(c, "failed to delete tyre record")
		return
	}
	h.logActivity(c, &vehicleID, userID, "deleted_tyre", "tyre_record", &id, "Deleted tyre record")
	response.OKMessage(c, "tyre record deleted")
}

func (h *TyreHandler) logActivity(c *gin.Context, vehicleID *uuid.UUID, userID uuid.UUID, action, entity string, entityID *uuid.UUID, description string) {
	log := &models.ActivityLog{
		ID: uuid.New(), VehicleID: vehicleID, UserID: userID,
		Action: action, Entity: &entity, EntityID: entityID, Description: &description, CreatedAt: time.Now(),
	}
	_ = h.activityRepo.Create(c.Request.Context(), log)
}

// ============================================================
// BATTERY HANDLER
// ============================================================

type BatteryHandler struct {
	batteryRepo  *repository.BatteryRepository
	activityRepo *repository.ActivityRepository
	validate     *validator.Validate
}

func NewBatteryHandler(batteryRepo *repository.BatteryRepository, activityRepo *repository.ActivityRepository) *BatteryHandler {
	return &BatteryHandler{batteryRepo: batteryRepo, activityRepo: activityRepo, validate: validator.New()}
}

func (h *BatteryHandler) List(c *gin.Context) {
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	records, err := h.batteryRepo.FindByVehicle(c.Request.Context(), vehicleID)
	if err != nil {
		response.InternalError(c, "failed to list battery records")
		return
	}
	response.OK(c, records)
}

func (h *BatteryHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))

	var req models.CreateBatteryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	now := time.Now()
	rec := &models.BatteryRecord{
		ID:               uuid.New(),
		VehicleID:        vehicleID,
		Brand:            req.Brand,
		Model:            req.Model,
		SerialNumber:     req.SerialNumber,
		InstallationDate: req.InstallationDate,
		InstallationKM:   req.InstallationKM,
		ReplacementDate:  req.ReplacementDate,
		ReplacementKM:    req.ReplacementKM,
		WarrantyExpiry:   req.WarrantyExpiry,
		Cost:             req.Cost,
		Remarks:          req.Remarks,
		CreatedBy:        userID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := h.batteryRepo.Create(c.Request.Context(), rec); err != nil {
		response.InternalError(c, "failed to create battery record")
		return
	}
	h.logActivity(c, &vehicleID, userID, "added_battery", "battery_record", &rec.ID, "Added battery record")
	response.Created(c, rec)
}

func (h *BatteryHandler) Get(c *gin.Context) {
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, _ := uuid.Parse(c.Param("id"))
	rec, err := h.batteryRepo.FindByID(c.Request.Context(), id, vehicleID)
	if err != nil || rec == nil {
		response.NotFound(c, "battery record not found")
		return
	}
	response.OK(c, rec)
}

func (h *BatteryHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, _ := uuid.Parse(c.Param("id"))

	rec, err := h.batteryRepo.FindByID(c.Request.Context(), id, vehicleID)
	if err != nil || rec == nil {
		response.NotFound(c, "battery record not found")
		return
	}

	var req models.UpdateBatteryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Brand != nil {
		rec.Brand = req.Brand
	}
	if req.Model != nil {
		rec.Model = req.Model
	}
	if req.SerialNumber != nil {
		rec.SerialNumber = req.SerialNumber
	}
	if req.InstallationDate != nil {
		rec.InstallationDate = req.InstallationDate
	}
	if req.InstallationKM != nil {
		rec.InstallationKM = req.InstallationKM
	}
	if req.ReplacementDate != nil {
		rec.ReplacementDate = req.ReplacementDate
	}
	if req.ReplacementKM != nil {
		rec.ReplacementKM = req.ReplacementKM
	}
	if req.WarrantyExpiry != nil {
		rec.WarrantyExpiry = req.WarrantyExpiry
	}
	if req.Cost != nil {
		rec.Cost = req.Cost
	}
	if req.Remarks != nil {
		rec.Remarks = req.Remarks
	}
	rec.UpdatedAt = time.Now()

	if err := h.batteryRepo.Update(c.Request.Context(), rec); err != nil {
		response.InternalError(c, "failed to update battery record")
		return
	}
	h.logActivity(c, &vehicleID, userID, "updated_battery", "battery_record", &rec.ID, "Updated battery record")
	response.OK(c, rec)
}

func (h *BatteryHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.batteryRepo.Delete(c.Request.Context(), id, vehicleID); err != nil {
		response.InternalError(c, "failed to delete battery record")
		return
	}
	h.logActivity(c, &vehicleID, userID, "deleted_battery", "battery_record", &id, "Deleted battery record")
	response.OKMessage(c, "battery record deleted")
}

func (h *BatteryHandler) logActivity(c *gin.Context, vehicleID *uuid.UUID, userID uuid.UUID, action, entity string, entityID *uuid.UUID, description string) {
	log := &models.ActivityLog{
		ID: uuid.New(), VehicleID: vehicleID, UserID: userID,
		Action: action, Entity: &entity, EntityID: entityID, Description: &description, CreatedAt: time.Now(),
	}
	_ = h.activityRepo.Create(c.Request.Context(), log)
}

// ============================================================
// ODOMETER HANDLER
// ============================================================

type OdometerHandler struct {
	odometerRepo *repository.OdometerRepository
	vehicleRepo  *repository.VehicleRepository
	activityRepo *repository.ActivityRepository
	validate     *validator.Validate
}

func NewOdometerHandler(odometerRepo *repository.OdometerRepository, vehicleRepo *repository.VehicleRepository, activityRepo *repository.ActivityRepository) *OdometerHandler {
	return &OdometerHandler{odometerRepo: odometerRepo, vehicleRepo: vehicleRepo, activityRepo: activityRepo, validate: validator.New()}
}

func (h *OdometerHandler) List(c *gin.Context) {
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	page, limit, offset := utils.GetPagination(c)
	records, total, err := h.odometerRepo.FindByVehicle(c.Request.Context(), vehicleID, limit, offset)
	if err != nil {
		response.InternalError(c, "failed to list odometer records")
		return
	}
	response.Paginated(c, records, total, page, limit)
}

func (h *OdometerHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))

	var req models.CreateOdometerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	now := time.Now()
	recordedAt := now
	if req.RecordedAt != nil {
		recordedAt = *req.RecordedAt
	}

	rec := &models.OdometerRecord{
		ID:         uuid.New(),
		VehicleID:  vehicleID,
		Reading:    req.Reading,
		RecordedAt: recordedAt,
		Source:     req.Source,
		Notes:      req.Notes,
		CreatedBy:  userID,
		CreatedAt:  now,
	}

	if err := h.odometerRepo.Create(c.Request.Context(), rec); err != nil {
		response.InternalError(c, "failed to create odometer record")
		return
	}

	_ = h.vehicleRepo.UpdateOdometer(c.Request.Context(), vehicleID, req.Reading)

	desc := fmt.Sprintf("Updated odometer to %.0f KM", req.Reading)
	log := &models.ActivityLog{
		ID: uuid.New(), VehicleID: &vehicleID, UserID: userID,
		Action: "updated_odometer", Entity: utils.StringPtr("odometer_record"), EntityID: &rec.ID,
		Description: &desc, CreatedAt: now,
	}
	_ = h.activityRepo.Create(c.Request.Context(), log)

	response.Created(c, rec)
}

// ============================================================
// REMINDER HANDLER
// ============================================================

type ReminderHandler struct {
	reminderRepo *repository.ReminderRepository
	activityRepo *repository.ActivityRepository
	validate     *validator.Validate
}

func NewReminderHandler(reminderRepo *repository.ReminderRepository, activityRepo *repository.ActivityRepository) *ReminderHandler {
	return &ReminderHandler{reminderRepo: reminderRepo, activityRepo: activityRepo, validate: validator.New()}
}

func (h *ReminderHandler) List(c *gin.Context) {
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	reminders, err := h.reminderRepo.FindByVehicle(c.Request.Context(), vehicleID)
	if err != nil {
		response.InternalError(c, "failed to list reminders")
		return
	}
	response.OK(c, reminders)
}

func (h *ReminderHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))

	var req models.CreateReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	now := time.Now()
	rem := &models.Reminder{
		ID:                uuid.New(),
		VehicleID:         vehicleID,
		Title:             req.Title,
		ReminderType:      req.ReminderType,
		DueDate:           req.DueDate,
		DueOdometer:       req.DueOdometer,
		AdvanceDays:       req.AdvanceDays,
		AdvanceKM:         req.AdvanceKM,
		RelatedEntityType: req.RelatedEntityType,
		RelatedEntityID:   req.RelatedEntityID,
		IsActive:          true,
		CreatedBy:         userID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := h.reminderRepo.Create(c.Request.Context(), rem); err != nil {
		response.InternalError(c, "failed to create reminder")
		return
	}

	desc := "Created reminder: " + rem.Title
	log := &models.ActivityLog{
		ID: uuid.New(), VehicleID: &vehicleID, UserID: userID,
		Action: "created_reminder", Entity: utils.StringPtr("reminder"), EntityID: &rem.ID,
		Description: &desc, CreatedAt: now,
	}
	_ = h.activityRepo.Create(c.Request.Context(), log)

	response.Created(c, rem)
}

func (h *ReminderHandler) Get(c *gin.Context) {
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, _ := uuid.Parse(c.Param("id"))
	rem, err := h.reminderRepo.FindByID(c.Request.Context(), id, vehicleID)
	if err != nil || rem == nil {
		response.NotFound(c, "reminder not found")
		return
	}
	response.OK(c, rem)
}

func (h *ReminderHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, _ := uuid.Parse(c.Param("id"))

	rem, err := h.reminderRepo.FindByID(c.Request.Context(), id, vehicleID)
	if err != nil || rem == nil {
		response.NotFound(c, "reminder not found")
		return
	}

	var req models.UpdateReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Title != nil {
		rem.Title = *req.Title
	}
	if req.DueDate != nil {
		rem.DueDate = req.DueDate
	}
	if req.DueOdometer != nil {
		rem.DueOdometer = req.DueOdometer
	}
	if req.AdvanceDays != nil {
		rem.AdvanceDays = req.AdvanceDays
	}
	if req.AdvanceKM != nil {
		rem.AdvanceKM = req.AdvanceKM
	}
	if req.IsActive != nil {
		rem.IsActive = *req.IsActive
	}
	rem.UpdatedAt = time.Now()

	if err := h.reminderRepo.Update(c.Request.Context(), rem); err != nil {
		response.InternalError(c, "failed to update reminder")
		return
	}
	_ = userID
	response.OK(c, rem)
}

func (h *ReminderHandler) Delete(c *gin.Context) {
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.reminderRepo.Delete(c.Request.Context(), id, vehicleID); err != nil {
		response.InternalError(c, "failed to delete reminder")
		return
	}
	response.OKMessage(c, "reminder deleted")
}

// ============================================================
// NOTIFICATION HANDLER
// ============================================================

type NotificationHandler struct {
	notificationRepo *repository.NotificationRepository
}

func NewNotificationHandler(notificationRepo *repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{notificationRepo: notificationRepo}
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	page, limit, offset := utils.GetPagination(c)
	notifications, total, err := h.notificationRepo.FindByUser(c.Request.Context(), userID, limit, offset)
	if err != nil {
		response.InternalError(c, "failed to list notifications")
		return
	}
	response.Paginated(c, notifications, total, page, limit)
}

// ============================================================
// DOCUMENT HANDLER
// ============================================================

type DocumentHandler struct {
	documentRepo *repository.DocumentRepository
	activityRepo *repository.ActivityRepository
	cfg          *config.Config
}

func NewDocumentHandler(documentRepo *repository.DocumentRepository, activityRepo *repository.ActivityRepository, cfg *config.Config) *DocumentHandler {
	return &DocumentHandler{documentRepo: documentRepo, activityRepo: activityRepo, cfg: cfg}
}

func (h *DocumentHandler) List(c *gin.Context) {
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	docs, err := h.documentRepo.FindByVehicle(c.Request.Context(), vehicleID)
	if err != nil {
		response.InternalError(c, "failed to list documents")
		return
	}
	for i := range docs {
		docs[i].FileURL = h.cfg.Storage.BaseURL + "/" + docs[i].FilePath
	}
	response.OK(c, docs)
}

func (h *DocumentHandler) Upload(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	newFileName := uuid.New().String() + ext
	uploadDir := filepath.Join(h.cfg.Storage.LocalPath, vehicleID.String())

	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		response.InternalError(c, "failed to create upload directory")
		return
	}

	destPath := filepath.Join(uploadDir, newFileName)
	dest, err := os.Create(destPath)
	if err != nil {
		response.InternalError(c, "failed to save file")
		return
	}
	defer dest.Close()

	size, err := io.Copy(dest, file)
	if err != nil {
		response.InternalError(c, "failed to write file")
		return
	}

	relPath := vehicleID.String() + "/" + newFileName
	mimeType := header.Header.Get("Content-Type")
	docType := c.PostForm("document_type")

	now := time.Now()
	doc := &models.Document{
		ID:           uuid.New(),
		VehicleID:    vehicleID,
		DocumentType: &docType,
		FileName:     header.Filename,
		FilePath:     relPath,
		FileSize:     &size,
		MimeType:     &mimeType,
		UploadedBy:   userID,
		CreatedAt:    now,
	}

	if entityType := c.PostForm("related_entity_type"); entityType != "" {
		doc.RelatedEntityType = &entityType
	}
	if entityIDStr := c.PostForm("related_entity_id"); entityIDStr != "" {
		if eid, err := uuid.Parse(entityIDStr); err == nil {
			doc.RelatedEntityID = &eid
		}
	}

	if err := h.documentRepo.Create(c.Request.Context(), doc); err != nil {
		response.InternalError(c, "failed to save document record")
		return
	}

	desc := "Uploaded document: " + header.Filename
	log := &models.ActivityLog{
		ID: uuid.New(), VehicleID: &vehicleID, UserID: userID,
		Action: "uploaded_document", Entity: utils.StringPtr("document"), EntityID: &doc.ID,
		Description: &desc, CreatedAt: now,
	}
	_ = h.activityRepo.Create(c.Request.Context(), log)

	doc.FileURL = h.cfg.Storage.BaseURL + "/" + relPath
	response.Created(c, doc)
}

func (h *DocumentHandler) Get(c *gin.Context) {
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, _ := uuid.Parse(c.Param("id"))
	doc, err := h.documentRepo.FindByID(c.Request.Context(), id, vehicleID)
	if err != nil || doc == nil {
		response.NotFound(c, "document not found")
		return
	}
	doc.FileURL = h.cfg.Storage.BaseURL + "/" + doc.FilePath
	response.OK(c, doc)
}

func (h *DocumentHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, _ := uuid.Parse(c.Param("id"))

	doc, err := h.documentRepo.FindByID(c.Request.Context(), id, vehicleID)
	if err != nil || doc == nil {
		response.NotFound(c, "document not found")
		return
	}

	fullPath := filepath.Join(h.cfg.Storage.LocalPath, doc.FilePath)
	_ = os.Remove(fullPath)

	if err := h.documentRepo.Delete(c.Request.Context(), id, vehicleID); err != nil {
		response.InternalError(c, "failed to delete document")
		return
	}

	desc := "Deleted document: " + doc.FileName
	log := &models.ActivityLog{
		ID: uuid.New(), VehicleID: &vehicleID, UserID: userID,
		Action: "deleted_document", Entity: utils.StringPtr("document"), EntityID: &id,
		Description: &desc, CreatedAt: time.Now(),
	}
	_ = h.activityRepo.Create(c.Request.Context(), log)

	response.OKMessage(c, "document deleted")
}

// ============================================================
// EXPENSE HANDLER
// ============================================================

type ExpenseHandler struct {
	expenseRepo  *repository.ExpenseRepository
	activityRepo *repository.ActivityRepository
	validate     *validator.Validate
}

func NewExpenseHandler(expenseRepo *repository.ExpenseRepository, activityRepo *repository.ActivityRepository) *ExpenseHandler {
	return &ExpenseHandler{expenseRepo: expenseRepo, activityRepo: activityRepo, validate: validator.New()}
}

func (h *ExpenseHandler) List(c *gin.Context) {
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	page, limit, offset := utils.GetPagination(c)
	expenses, total, err := h.expenseRepo.FindByVehicle(c.Request.Context(), vehicleID, limit, offset)
	if err != nil {
		response.InternalError(c, "failed to list expenses")
		return
	}
	response.Paginated(c, expenses, total, page, limit)
}

func (h *ExpenseHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))

	var req models.CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	now := time.Now()
	exp := &models.Expense{
		ID:          uuid.New(),
		VehicleID:   vehicleID,
		Category:    req.Category,
		ExpenseDate: req.ExpenseDate,
		Description: req.Description,
		Amount:      req.Amount,
		Vendor:      req.Vendor,
		Odometer:    req.Odometer,
		Notes:       req.Notes,
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.expenseRepo.Create(c.Request.Context(), exp); err != nil {
		response.InternalError(c, "failed to create expense")
		return
	}

	desc := fmt.Sprintf("Added expense: %s (%.2f)", req.Category, req.Amount)
	log := &models.ActivityLog{
		ID: uuid.New(), VehicleID: &vehicleID, UserID: userID,
		Action: "added_expense", Entity: utils.StringPtr("expense"), EntityID: &exp.ID,
		Description: &desc, CreatedAt: now,
	}
	_ = h.activityRepo.Create(c.Request.Context(), log)

	response.Created(c, exp)
}

func (h *ExpenseHandler) Get(c *gin.Context) {
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, _ := uuid.Parse(c.Param("id"))
	exp, err := h.expenseRepo.FindByID(c.Request.Context(), id, vehicleID)
	if err != nil || exp == nil {
		response.NotFound(c, "expense not found")
		return
	}
	response.OK(c, exp)
}

func (h *ExpenseHandler) Update(c *gin.Context) {
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, _ := uuid.Parse(c.Param("id"))

	exp, err := h.expenseRepo.FindByID(c.Request.Context(), id, vehicleID)
	if err != nil || exp == nil {
		response.NotFound(c, "expense not found")
		return
	}

	var req models.UpdateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Category != nil {
		exp.Category = *req.Category
	}
	if req.ExpenseDate != nil {
		exp.ExpenseDate = *req.ExpenseDate
	}
	if req.Description != nil {
		exp.Description = req.Description
	}
	if req.Amount != nil {
		exp.Amount = *req.Amount
	}
	if req.Vendor != nil {
		exp.Vendor = req.Vendor
	}
	if req.Odometer != nil {
		exp.Odometer = req.Odometer
	}
	if req.Notes != nil {
		exp.Notes = req.Notes
	}
	exp.UpdatedAt = time.Now()

	if err := h.expenseRepo.Update(c.Request.Context(), exp); err != nil {
		response.InternalError(c, "failed to update expense")
		return
	}
	response.OK(c, exp)
}

func (h *ExpenseHandler) Delete(c *gin.Context) {
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.expenseRepo.Delete(c.Request.Context(), id, vehicleID); err != nil {
		response.InternalError(c, "failed to delete expense")
		return
	}
	response.OKMessage(c, "expense deleted")
}

// ============================================================
// ACTIVITY HANDLER
// ============================================================

type ActivityHandler struct {
	activityRepo *repository.ActivityRepository
}

func NewActivityHandler(activityRepo *repository.ActivityRepository) *ActivityHandler {
	return &ActivityHandler{activityRepo: activityRepo}
}

func (h *ActivityHandler) List(c *gin.Context) {
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))
	page, limit, offset := utils.GetPagination(c)
	logs, total, err := h.activityRepo.FindByVehicle(c.Request.Context(), vehicleID, limit, offset)
	if err != nil {
		response.InternalError(c, "failed to list activity logs")
		return
	}
	response.Paginated(c, logs, total, page, limit)
}

// ============================================================
// REPORT HANDLER
// ============================================================

type ReportHandler struct {
	vehicleRepo  *repository.VehicleRepository
	expenseRepo  *repository.ExpenseRepository
	insuranceRepo *repository.InsuranceRepository
	pucRepo      *repository.PUCRepository
	serviceRepo  *repository.ServiceRepository
	reminderRepo *repository.ReminderRepository
	activityRepo *repository.ActivityRepository
	odometerRepo *repository.OdometerRepository
}

func NewReportHandler(
	vehicleRepo *repository.VehicleRepository,
	expenseRepo *repository.ExpenseRepository,
	insuranceRepo *repository.InsuranceRepository,
	pucRepo *repository.PUCRepository,
	serviceRepo *repository.ServiceRepository,
	reminderRepo *repository.ReminderRepository,
	activityRepo *repository.ActivityRepository,
	odometerRepo *repository.OdometerRepository,
) *ReportHandler {
	return &ReportHandler{
		vehicleRepo:   vehicleRepo,
		expenseRepo:   expenseRepo,
		insuranceRepo: insuranceRepo,
		pucRepo:       pucRepo,
		serviceRepo:   serviceRepo,
		reminderRepo:  reminderRepo,
		activityRepo:  activityRepo,
		odometerRepo:  odometerRepo,
	}
}

func (h *ReportHandler) VehicleSummary(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))

	vehicle, err := h.vehicleRepo.FindByIDWithRole(c.Request.Context(), vehicleID, userID)
	if err != nil || vehicle == nil {
		response.NotFound(c, "vehicle not found")
		return
	}

	insurance, _ := h.insuranceRepo.GetLatest(c.Request.Context(), vehicleID)
	puc, _ := h.pucRepo.GetLatest(c.Request.Context(), vehicleID)
	lastService, _ := h.serviceRepo.GetLatest(c.Request.Context(), vehicleID)
	recentActivity, _ := h.activityRepo.FindRecentByVehicle(c.Request.Context(), vehicleID, 10)
	reminders, _ := h.reminderRepo.FindByVehicle(c.Request.Context(), vehicleID)

	reminderStatuses := buildReminderStatuses(reminders, vehicle.CurrentOdometer)

	if insurance != nil {
		insurance.Status = insuranceStatus(insurance.ExpiryDate)
	}
	if puc != nil {
		puc.Status = insuranceStatus(puc.ExpiryDate)
	}

	summary := gin.H{
		"vehicle":            vehicle,
		"latest_insurance":   insurance,
		"latest_puc":         puc,
		"last_service":       lastService,
		"upcoming_reminders": reminderStatuses,
		"recent_activity":    recentActivity,
	}

	response.OK(c, summary)
}

func (h *ReportHandler) ExpenseReport(c *gin.Context) {
	vehicleID, _ := uuid.Parse(c.Param("vehicleId"))

	byCategory, err := h.expenseRepo.GetSummary(c.Request.Context(), vehicleID)
	if err != nil {
		response.InternalError(c, "failed to generate expense report")
		return
	}

	now := time.Now()
	startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	annual, _ := h.expenseRepo.GetTotalByPeriod(c.Request.Context(), vehicleID, startOfYear, now)
	monthly, _ := h.expenseRepo.GetTotalByPeriod(c.Request.Context(), vehicleID, startOfMonth, now)

	var total float64
	for _, v := range byCategory {
		total += v
	}

	report := models.ExpenseReport{
		TotalExpense:      total,
		AnnualExpense:     annual,
		MonthlyExpense:    monthly,
		ExpenseByCategory: byCategory,
	}

	vehicle, _ := h.vehicleRepo.FindByID(c.Request.Context(), vehicleID)
	if vehicle != nil && vehicle.CurrentOdometer > 0 && total > 0 {
		latestOdometer, _ := h.odometerRepo.GetLatest(c.Request.Context(), vehicleID)
		if latestOdometer != nil && latestOdometer.Reading > 0 {
			costPerKM := total / latestOdometer.Reading
			report.CostPerKM = &costPerKM
		}
	}

	response.OK(c, report)
}

func (h *ReportHandler) UserDashboard(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	vehicles, err := h.vehicleRepo.FindAllByUserID(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "failed to get dashboard")
		return
	}

	var allReminderStatuses []models.ReminderStatus
	for _, v := range vehicles {
		reminders, _ := h.reminderRepo.FindByVehicle(c.Request.Context(), v.ID)
		statuses := buildReminderStatuses(reminders, v.CurrentOdometer)
		allReminderStatuses = append(allReminderStatuses, statuses...)
	}

	dashboard := models.UserDashboard{
		TotalVehicles:     len(vehicles),
		Vehicles:          vehicles,
		UpcomingReminders: allReminderStatuses,
	}

	response.OK(c, dashboard)
}

func buildReminderStatuses(reminders []models.Reminder, currentKM float64) []models.ReminderStatus {
	var statuses []models.ReminderStatus
	now := time.Now()

	for i := range reminders {
		rem := &reminders[i]
		if !rem.IsActive {
			continue
		}

		status := models.ReminderStatus{Reminder: rem, Status: "normal"}

		switch rem.ReminderType {
		case "date":
			if rem.DueDate != nil {
				daysLeft := int(rem.DueDate.Sub(now).Hours() / 24)
				status.DaysLeft = &daysLeft
				threshold := 30
				if rem.AdvanceDays != nil {
					threshold = *rem.AdvanceDays
				}
				if daysLeft < 0 {
					status.Status = "overdue"
				} else if daysLeft <= threshold {
					status.Status = "due_soon"
				}
			}
		case "odometer":
			if rem.DueOdometer != nil {
				kmLeft := *rem.DueOdometer - currentKM
				status.KMLeft = &kmLeft
				threshold := 1000.0
				if rem.AdvanceKM != nil {
					threshold = *rem.AdvanceKM
				}
				if kmLeft <= 0 {
					status.Status = "overdue"
				} else if kmLeft <= threshold {
					status.Status = "due_soon"
				}
			}
		case "date_or_odometer":
			dateOverdue, kmOverdue := false, false
			if rem.DueDate != nil {
				daysLeft := int(rem.DueDate.Sub(now).Hours() / 24)
				status.DaysLeft = &daysLeft
				threshold := 30
				if rem.AdvanceDays != nil {
					threshold = *rem.AdvanceDays
				}
				if daysLeft < 0 {
					dateOverdue = true
				} else if daysLeft <= threshold {
					status.Status = "due_soon"
				}
			}
			if rem.DueOdometer != nil {
				kmLeft := *rem.DueOdometer - currentKM
				status.KMLeft = &kmLeft
				threshold := 1000.0
				if rem.AdvanceKM != nil {
					threshold = *rem.AdvanceKM
				}
				if kmLeft <= 0 {
					kmOverdue = true
				} else if kmLeft <= threshold && status.Status == "normal" {
					status.Status = "due_soon"
				}
			}
			if dateOverdue || kmOverdue {
				status.Status = "overdue"
			}
		}

		statuses = append(statuses, status)
	}

	if statuses == nil {
		statuses = []models.ReminderStatus{}
	}

	return statuses
}

func buildDocumentPath(vehicleID, fileName string) string {
	return strings.Join([]string{vehicleID, fileName}, "/")
}
