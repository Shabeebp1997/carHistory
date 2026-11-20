package models

import (
	"time"

	"github.com/google/uuid"
)

type VehicleType struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Vehicle struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	VehicleTypeID      int        `json:"vehicle_type_id" db:"vehicle_type_id"`
	VehicleTypeName    string     `json:"vehicle_type_name,omitempty" db:"vehicle_type_name"`
	Brand              *string    `json:"brand" db:"brand"`
	Model              *string    `json:"model" db:"model"`
	Variant            *string    `json:"variant" db:"variant"`
	ManufacturingYear  *int       `json:"manufacturing_year" db:"manufacturing_year"`
	PurchaseDate       *time.Time `json:"purchase_date" db:"purchase_date"`
	RegistrationNumber *string    `json:"registration_number" db:"registration_number"`
	ChassisNumber      *string    `json:"chassis_number" db:"chassis_number"`
	EngineNumber       *string    `json:"engine_number" db:"engine_number"`
	RCOwner            *string    `json:"rc_owner" db:"rc_owner"`
	OwnerContact       *string    `json:"owner_contact" db:"owner_contact"`
	FuelType           *string    `json:"fuel_type" db:"fuel_type"`
	CurrentOdometer    float64    `json:"current_odometer" db:"current_odometer"`
	Colour             *string    `json:"colour" db:"colour"`
	Nickname           *string    `json:"nickname" db:"nickname"`
	PurchasePrice      *float64   `json:"purchase_price" db:"purchase_price"`
	Notes              *string    `json:"notes" db:"notes"`
	IsArchived         bool       `json:"is_archived" db:"is_archived"`
	CreatedBy          uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
	UserRole           string     `json:"user_role,omitempty" db:"user_role"`
}

type VehicleMember struct {
	ID        uuid.UUID `json:"id" db:"id"`
	VehicleID uuid.UUID `json:"vehicle_id" db:"vehicle_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	UserName  string    `json:"user_name,omitempty" db:"user_name"`
	UserEmail string    `json:"user_email,omitempty" db:"user_email"`
	Role      string    `json:"role" db:"role"`
	JoinedAt  time.Time `json:"joined_at" db:"joined_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type VehicleInvitation struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	VehicleID    uuid.UUID  `json:"vehicle_id" db:"vehicle_id"`
	VehicleName  string     `json:"vehicle_name,omitempty" db:"vehicle_name"`
	InvitedBy    uuid.UUID  `json:"invited_by" db:"invited_by"`
	InviterName  string     `json:"inviter_name,omitempty" db:"inviter_name"`
	InvitedEmail string     `json:"invited_email" db:"invited_email"`
	Role         string     `json:"role" db:"role"`
	Token        string     `json:"token,omitempty" db:"token"`
	Status       string     `json:"status" db:"status"`
	ExpiresAt    time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

type CreateVehicleRequest struct {
	VehicleTypeID      int        `json:"vehicle_type_id" validate:"required"`
	Brand              *string    `json:"brand"`
	Model              *string    `json:"model"`
	Variant            *string    `json:"variant"`
	ManufacturingYear  *int       `json:"manufacturing_year"`
	PurchaseDate       *time.Time `json:"purchase_date"`
	RegistrationNumber *string    `json:"registration_number"`
	ChassisNumber      *string    `json:"chassis_number"`
	EngineNumber       *string    `json:"engine_number"`
	RCOwner            *string    `json:"rc_owner"`
	OwnerContact       *string    `json:"owner_contact"`
	FuelType           *string    `json:"fuel_type"`
	CurrentOdometer    float64    `json:"current_odometer"`
	Colour             *string    `json:"colour"`
	Nickname           *string    `json:"nickname"`
	PurchasePrice      *float64   `json:"purchase_price"`
	Notes              *string    `json:"notes"`
}

type UpdateVehicleRequest struct {
	VehicleTypeID      *int       `json:"vehicle_type_id"`
	Brand              *string    `json:"brand"`
	Model              *string    `json:"model"`
	Variant            *string    `json:"variant"`
	ManufacturingYear  *int       `json:"manufacturing_year"`
	PurchaseDate       *time.Time `json:"purchase_date"`
	RegistrationNumber *string    `json:"registration_number"`
	ChassisNumber      *string    `json:"chassis_number"`
	EngineNumber       *string    `json:"engine_number"`
	RCOwner            *string    `json:"rc_owner"`
	OwnerContact       *string    `json:"owner_contact"`
	FuelType           *string    `json:"fuel_type"`
	CurrentOdometer    *float64   `json:"current_odometer"`
	Colour             *string    `json:"colour"`
	Nickname           *string    `json:"nickname"`
	PurchasePrice      *float64   `json:"purchase_price"`
	Notes              *string    `json:"notes"`
}

type InviteMemberRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=editor viewer"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=editor viewer"`
}
