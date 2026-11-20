package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"carhistory/internal/config"
	"carhistory/internal/models"
	"carhistory/internal/repository"
	"carhistory/pkg/utils"
)

var (
	ErrVehicleNotFound   = errors.New("vehicle not found")
	ErrAccessDenied      = errors.New("access denied")
	ErrAlreadyMember     = errors.New("user is already a member of this vehicle")
	ErrCannotRemoveOwner = errors.New("cannot remove the vehicle owner")
	ErrInvitationInvalid = errors.New("invitation is invalid or has expired")
)

type VehicleService struct {
	vehicleRepo  *repository.VehicleRepository
	activityRepo *repository.ActivityRepository
	userRepo     *repository.UserRepository
	emailSvc     *EmailService
	cfg          *config.Config
}

func NewVehicleService(
	vehicleRepo *repository.VehicleRepository,
	activityRepo *repository.ActivityRepository,
	userRepo *repository.UserRepository,
	emailSvc *EmailService,
	cfg *config.Config,
) *VehicleService {
	return &VehicleService{
		vehicleRepo:  vehicleRepo,
		activityRepo: activityRepo,
		userRepo:     userRepo,
		emailSvc:     emailSvc,
		cfg:          cfg,
	}
}

func (s *VehicleService) CreateVehicle(ctx context.Context, userID uuid.UUID, req *models.CreateVehicleRequest) (*models.Vehicle, error) {
	now := time.Now()
	vehicle := &models.Vehicle{
		ID:                 uuid.New(),
		VehicleTypeID:      req.VehicleTypeID,
		Brand:              req.Brand,
		Model:              req.Model,
		Variant:            req.Variant,
		ManufacturingYear:  req.ManufacturingYear,
		PurchaseDate:       req.PurchaseDate,
		RegistrationNumber: req.RegistrationNumber,
		ChassisNumber:      req.ChassisNumber,
		EngineNumber:       req.EngineNumber,
		RCOwner:            req.RCOwner,
		OwnerContact:       req.OwnerContact,
		FuelType:           req.FuelType,
		CurrentOdometer:    req.CurrentOdometer,
		Colour:             req.Colour,
		Nickname:           req.Nickname,
		Notes:              req.Notes,
		IsArchived:         false,
		CreatedBy:          userID,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := s.vehicleRepo.Create(ctx, vehicle); err != nil {
		return nil, err
	}

	member := &models.VehicleMember{
		ID:        uuid.New(),
		VehicleID: vehicle.ID,
		UserID:    userID,
		Role:      "owner",
		JoinedAt:  now,
		CreatedAt: now,
	}
	if err := s.vehicleRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}

	s.logActivity(ctx, &vehicle.ID, userID, "created_vehicle", "vehicle", &vehicle.ID,
		"Vehicle created")

	return s.vehicleRepo.FindByIDWithRole(ctx, vehicle.ID, userID)
}

func (s *VehicleService) GetVehicle(ctx context.Context, vehicleID, userID uuid.UUID) (*models.Vehicle, error) {
	vehicle, err := s.vehicleRepo.FindByIDWithRole(ctx, vehicleID, userID)
	if err != nil {
		return nil, err
	}
	if vehicle == nil {
		return nil, ErrVehicleNotFound
	}
	return vehicle, nil
}

func (s *VehicleService) ListVehicles(ctx context.Context, userID uuid.UUID) ([]models.Vehicle, error) {
	return s.vehicleRepo.FindAllByUserID(ctx, userID)
}

func (s *VehicleService) UpdateVehicle(ctx context.Context, vehicleID, userID uuid.UUID, req *models.UpdateVehicleRequest) (*models.Vehicle, error) {
	vehicle, err := s.vehicleRepo.FindByID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	if vehicle == nil {
		return nil, ErrVehicleNotFound
	}

	if req.VehicleTypeID != nil {
		vehicle.VehicleTypeID = *req.VehicleTypeID
	}
	if req.Brand != nil {
		vehicle.Brand = req.Brand
	}
	if req.Model != nil {
		vehicle.Model = req.Model
	}
	if req.Variant != nil {
		vehicle.Variant = req.Variant
	}
	if req.ManufacturingYear != nil {
		vehicle.ManufacturingYear = req.ManufacturingYear
	}
	if req.PurchaseDate != nil {
		vehicle.PurchaseDate = req.PurchaseDate
	}
	if req.RegistrationNumber != nil {
		vehicle.RegistrationNumber = req.RegistrationNumber
	}
	if req.ChassisNumber != nil {
		vehicle.ChassisNumber = req.ChassisNumber
	}
	if req.EngineNumber != nil {
		vehicle.EngineNumber = req.EngineNumber
	}
	if req.RCOwner != nil {
		vehicle.RCOwner = req.RCOwner
	}
	if req.OwnerContact != nil {
		vehicle.OwnerContact = req.OwnerContact
	}
	if req.FuelType != nil {
		vehicle.FuelType = req.FuelType
	}
	if req.CurrentOdometer != nil {
		vehicle.CurrentOdometer = *req.CurrentOdometer
	}
	if req.Colour != nil {
		vehicle.Colour = req.Colour
	}
	if req.Nickname != nil {
		vehicle.Nickname = req.Nickname
	}
	if req.Notes != nil {
		vehicle.Notes = req.Notes
	}
	vehicle.UpdatedAt = time.Now()

	if err := s.vehicleRepo.Update(ctx, vehicle); err != nil {
		return nil, err
	}

	s.logActivity(ctx, &vehicleID, userID, "updated_vehicle", "vehicle", &vehicleID, "Vehicle information updated")

	return s.vehicleRepo.FindByIDWithRole(ctx, vehicleID, userID)
}

func (s *VehicleService) ArchiveVehicle(ctx context.Context, vehicleID, userID uuid.UUID) error {
	if err := s.vehicleRepo.Archive(ctx, vehicleID); err != nil {
		return err
	}
	s.logActivity(ctx, &vehicleID, userID, "archived_vehicle", "vehicle", &vehicleID, "Vehicle archived")
	return nil
}

func (s *VehicleService) InviteMember(ctx context.Context, vehicleID, inviterID uuid.UUID, req *models.InviteMemberRequest) error {
	vehicle, err := s.vehicleRepo.FindByID(ctx, vehicleID)
	if err != nil || vehicle == nil {
		return ErrVehicleNotFound
	}

	invitee, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return err
	}

	if invitee != nil {
		existing, err := s.vehicleRepo.GetMember(ctx, vehicleID, invitee.ID)
		if err != nil {
			return err
		}
		if existing != nil {
			return ErrAlreadyMember
		}
	}

	token, err := utils.GenerateSecureToken(32)
	if err != nil {
		return err
	}

	now := time.Now()
	invitation := &models.VehicleInvitation{
		ID:           uuid.New(),
		VehicleID:    vehicleID,
		InvitedBy:    inviterID,
		InvitedEmail: req.Email,
		Role:         req.Role,
		Token:        token,
		Status:       "pending",
		ExpiresAt:    now.Add(time.Duration(s.cfg.App.InvitationExpiryHours) * time.Hour),
		CreatedAt:    now,
	}

	if err := s.vehicleRepo.CreateInvitation(ctx, invitation); err != nil {
		return err
	}

	inviter, _ := s.userRepo.FindByID(ctx, inviterID)
	inviterName := "Someone"
	if inviter != nil {
		inviterName = inviter.Name
	}

	vehicleName := "your vehicle"
	if vehicle.Nickname != nil {
		vehicleName = *vehicle.Nickname
	}

	acceptURL := s.cfg.App.BaseURL + "/invitations/" + token

	go s.emailSvc.SendVehicleInvitation(req.Email, inviterName, vehicleName, req.Role, acceptURL)

	s.logActivity(ctx, &vehicleID, inviterID, "invited_member", "invitation", &invitation.ID,
		"Invited "+req.Email+" as "+req.Role)

	return nil
}

func (s *VehicleService) AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) error {
	inv, err := s.vehicleRepo.FindInvitationByToken(ctx, token)
	if err != nil {
		return err
	}
	if inv == nil || inv.Status != "pending" || inv.ExpiresAt.Before(time.Now()) {
		return ErrInvitationInvalid
	}

	existing, err := s.vehicleRepo.GetMember(ctx, inv.VehicleID, userID)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrAlreadyMember
	}

	now := time.Now()
	member := &models.VehicleMember{
		ID:        uuid.New(),
		VehicleID: inv.VehicleID,
		UserID:    userID,
		Role:      inv.Role,
		JoinedAt:  now,
		CreatedAt: now,
	}
	if err := s.vehicleRepo.AddMember(ctx, member); err != nil {
		return err
	}

	if err := s.vehicleRepo.UpdateInvitationStatus(ctx, inv.ID, "accepted"); err != nil {
		return err
	}

	s.logActivity(ctx, &inv.VehicleID, userID, "accepted_invitation", "invitation", &inv.ID,
		"Joined as "+inv.Role)

	return nil
}

func (s *VehicleService) RejectInvitation(ctx context.Context, token string) error {
	inv, err := s.vehicleRepo.FindInvitationByToken(ctx, token)
	if err != nil {
		return err
	}
	if inv == nil || inv.Status != "pending" {
		return ErrInvitationInvalid
	}
	return s.vehicleRepo.UpdateInvitationStatus(ctx, inv.ID, "rejected")
}

func (s *VehicleService) RemoveMember(ctx context.Context, vehicleID, requesterID, targetUserID uuid.UUID) error {
	target, err := s.vehicleRepo.GetMember(ctx, vehicleID, targetUserID)
	if err != nil {
		return err
	}
	if target == nil {
		return errors.New("member not found")
	}
	if target.Role == "owner" {
		return ErrCannotRemoveOwner
	}

	if err := s.vehicleRepo.RemoveMember(ctx, vehicleID, targetUserID); err != nil {
		return err
	}

	s.logActivity(ctx, &vehicleID, requesterID, "removed_member", "vehicle_member", &target.ID,
		"Removed member from vehicle")
	return nil
}

func (s *VehicleService) UpdateMemberRole(ctx context.Context, vehicleID, requesterID, targetUserID uuid.UUID, role string) error {
	target, err := s.vehicleRepo.GetMember(ctx, vehicleID, targetUserID)
	if err != nil {
		return err
	}
	if target == nil {
		return errors.New("member not found")
	}
	if target.Role == "owner" {
		return errors.New("cannot change the owner's role")
	}

	if err := s.vehicleRepo.UpdateMemberRole(ctx, vehicleID, targetUserID, role); err != nil {
		return err
	}

	s.logActivity(ctx, &vehicleID, requesterID, "updated_member_role", "vehicle_member", &target.ID,
		"Updated member role to "+role)
	return nil
}

func (s *VehicleService) logActivity(ctx context.Context, vehicleID *uuid.UUID, userID uuid.UUID, action, entity string, entityID *uuid.UUID, description string) {
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
	_ = s.activityRepo.Create(ctx, log)
}
