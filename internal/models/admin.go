package models

import (
	"time"

	"github.com/google/uuid"
)

type AdminUserListItem struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Name           string     `json:"name" db:"name"`
	Email          *string    `json:"email" db:"email"`
	MobileNumber   *string    `json:"mobile_number" db:"mobile_number"`
	ProfilePicture *string    `json:"profile_picture" db:"profile_picture"`
	IsAdmin        bool       `json:"is_admin" db:"is_admin"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	VehicleCount   int        `json:"vehicle_count" db:"vehicle_count"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

type VehicleTypeCount struct {
	TypeName string `json:"type_name" db:"type_name"`
	Count    int    `json:"count" db:"count"`
}

type FuelTypeCount struct {
	FuelType string `json:"fuel_type" db:"fuel_type"`
	Count    int    `json:"count" db:"count"`
}

type BrandCount struct {
	Brand string `json:"brand" db:"brand"`
	Count int    `json:"count" db:"count"`
}

type ModelCount struct {
	Brand string `json:"brand" db:"brand"`
	Model string `json:"model" db:"model"`
	Count int    `json:"count" db:"count"`
}

type UserAnalytics struct {
	Total        int `json:"total"`
	Active       int `json:"active"`
	Admins       int `json:"admins"`
	WithEmail    int `json:"with_email"`
	WithPhone    int `json:"with_phone"`
	NewThisMonth int `json:"new_this_month"`
}

type VehicleAnalytics struct {
	Total          int                `json:"total"`
	Active         int                `json:"active"`
	ByType         []VehicleTypeCount `json:"by_type"`
	ByFuelType     []FuelTypeCount    `json:"by_fuel_type"`
	PopularBrands  []BrandCount       `json:"popular_brands"`
	PopularModels  []ModelCount       `json:"popular_models"`
}

type ActivityAnalytics struct {
	TotalMaintenanceRecords int `json:"total_maintenance_records"`
	TotalServiceRecords     int `json:"total_service_records"`
	TotalInsuranceRecords   int `json:"total_insurance_records"`
	TotalFuelLogs           int `json:"total_fuel_logs"`
	TotalDocuments          int `json:"total_documents"`
	TotalReminders          int `json:"total_reminders"`
	ActiveReminders         int `json:"active_reminders"`
	TotalMedia              int `json:"total_media"`
}

type AdminAnalytics struct {
	Users    UserAnalytics    `json:"users"`
	Vehicles VehicleAnalytics `json:"vehicles"`
	Activity ActivityAnalytics `json:"activity"`
}
