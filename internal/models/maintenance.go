package models

import (
	"time"

	"github.com/google/uuid"
)

type MaintenanceType struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description" db:"description"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type MaintenanceRecord struct {
	ID                    uuid.UUID  `json:"id" db:"id"`
	VehicleID             uuid.UUID  `json:"vehicle_id" db:"vehicle_id"`
	MaintenanceTypeID     int        `json:"maintenance_type_id" db:"maintenance_type_id"`
	MaintenanceTypeName   string     `json:"maintenance_type_name,omitempty" db:"maintenance_type_name"`
	Title                 string     `json:"title" db:"title"`
	DatePerformed         time.Time  `json:"date_performed" db:"date_performed"`
	OdometerAtMaintenance *float64   `json:"odometer_at_maintenance" db:"odometer_at_maintenance"`
	NextDueDate           *time.Time `json:"next_due_date" db:"next_due_date"`
	NextDueOdometer       *float64   `json:"next_due_odometer" db:"next_due_odometer"`
	Cost                  *float64   `json:"cost" db:"cost"`
	ServiceProvider       *string    `json:"service_provider" db:"service_provider"`
	Description           *string    `json:"description" db:"description"`
	Remarks               *string    `json:"remarks" db:"remarks"`
	CreatedBy             uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateMaintenanceRequest struct {
	MaintenanceTypeID     int        `json:"maintenance_type_id" validate:"required"`
	Title                 string     `json:"title" validate:"required,max=255"`
	DatePerformed         time.Time  `json:"date_performed" validate:"required"`
	OdometerAtMaintenance *float64   `json:"odometer_at_maintenance"`
	NextDueDate           *time.Time `json:"next_due_date"`
	NextDueOdometer       *float64   `json:"next_due_odometer"`
	Cost                  *float64   `json:"cost"`
	ServiceProvider       *string    `json:"service_provider"`
	Description           *string    `json:"description"`
	Remarks               *string    `json:"remarks"`
}

type UpdateMaintenanceRequest struct {
	MaintenanceTypeID     *int       `json:"maintenance_type_id"`
	Title                 *string    `json:"title"`
	DatePerformed         *time.Time `json:"date_performed"`
	OdometerAtMaintenance *float64   `json:"odometer_at_maintenance"`
	NextDueDate           *time.Time `json:"next_due_date"`
	NextDueOdometer       *float64   `json:"next_due_odometer"`
	Cost                  *float64   `json:"cost"`
	ServiceProvider       *string    `json:"service_provider"`
	Description           *string    `json:"description"`
	Remarks               *string    `json:"remarks"`
}

type ServiceRecord struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	VehicleID       uuid.UUID  `json:"vehicle_id" db:"vehicle_id"`
	ServiceDate     time.Time  `json:"service_date" db:"service_date"`
	Odometer        *float64   `json:"odometer" db:"odometer"`
	ServiceType     *string    `json:"service_type" db:"service_type"`
	ServiceCentre   *string    `json:"service_centre" db:"service_centre"`
	Technician      *string    `json:"technician" db:"technician"`
	LabourCost      float64    `json:"labour_cost" db:"labour_cost"`
	PartsCost       float64    `json:"parts_cost" db:"parts_cost"`
	OtherCost       float64    `json:"other_cost" db:"other_cost"`
	TotalCost       float64    `json:"total_cost" db:"total_cost"`
	NextServiceDate *time.Time `json:"next_service_date" db:"next_service_date"`
	NextServiceKM   *float64   `json:"next_service_km" db:"next_service_km"`
	Remarks         *string    `json:"remarks" db:"remarks"`
	Items           []ServiceItem `json:"items,omitempty"`
	CreatedBy       uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

type ServiceItem struct {
	ID              uuid.UUID `json:"id" db:"id"`
	ServiceRecordID uuid.UUID `json:"service_record_id" db:"service_record_id"`
	Name            string    `json:"name" db:"name"`
	Description     *string   `json:"description" db:"description"`
	Cost            float64   `json:"cost" db:"cost"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type ServiceItemInput struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description"`
	Cost        float64 `json:"cost"`
}

type CreateServiceRequest struct {
	ServiceDate     time.Time          `json:"service_date" validate:"required"`
	Odometer        *float64           `json:"odometer"`
	ServiceType     *string            `json:"service_type"`
	ServiceCentre   *string            `json:"service_centre"`
	Technician      *string            `json:"technician"`
	LabourCost      float64            `json:"labour_cost"`
	PartsCost       float64            `json:"parts_cost"`
	OtherCost       float64            `json:"other_cost"`
	NextServiceDate *time.Time         `json:"next_service_date"`
	NextServiceKM   *float64           `json:"next_service_km"`
	Remarks         *string            `json:"remarks"`
	Items           []ServiceItemInput `json:"items"`
}

type UpdateServiceRequest struct {
	ServiceDate     *time.Time         `json:"service_date"`
	Odometer        *float64           `json:"odometer"`
	ServiceType     *string            `json:"service_type"`
	ServiceCentre   *string            `json:"service_centre"`
	Technician      *string            `json:"technician"`
	LabourCost      *float64           `json:"labour_cost"`
	PartsCost       *float64           `json:"parts_cost"`
	OtherCost       *float64           `json:"other_cost"`
	NextServiceDate *time.Time         `json:"next_service_date"`
	NextServiceKM   *float64           `json:"next_service_km"`
	Remarks         *string            `json:"remarks"`
	Items           []ServiceItemInput `json:"items"`
}
