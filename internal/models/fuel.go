package models

import (
	"time"

	"github.com/google/uuid"
)

type FuelLog struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	VehicleID       uuid.UUID  `json:"vehicle_id" db:"vehicle_id"`
	LoggedBy        uuid.UUID  `json:"logged_by" db:"logged_by"`
	FuelType        string     `json:"fuel_type" db:"fuel_type"`
	Quantity        *float64   `json:"quantity" db:"quantity"`
	PricePerUnit    *float64   `json:"price_per_unit" db:"price_per_unit"`
	TotalCost       float64    `json:"total_cost" db:"total_cost"`
	OdometerReading *float64   `json:"odometer_reading" db:"odometer_reading"`
	FuelStation     *string    `json:"fuel_station" db:"fuel_station"`
	Date            time.Time  `json:"date" db:"date"`
	Notes           *string    `json:"notes" db:"notes"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateFuelLogRequest struct {
	FuelType        string     `json:"fuel_type" validate:"required,oneof=petrol diesel cng electric hybrid"`
	Quantity        *float64   `json:"quantity"`
	PricePerUnit    *float64   `json:"price_per_unit"`
	TotalCost       float64    `json:"total_cost" validate:"required,gt=0"`
	OdometerReading *float64   `json:"odometer_reading"`
	FuelStation     *string    `json:"fuel_station"`
	Date            time.Time  `json:"date" validate:"required"`
	Notes           *string    `json:"notes"`
}

type UpdateFuelLogRequest struct {
	FuelType        *string    `json:"fuel_type" validate:"omitempty,oneof=petrol diesel cng electric hybrid"`
	Quantity        *float64   `json:"quantity"`
	PricePerUnit    *float64   `json:"price_per_unit"`
	TotalCost       *float64   `json:"total_cost" validate:"omitempty,gt=0"`
	OdometerReading *float64   `json:"odometer_reading"`
	FuelStation     *string    `json:"fuel_station"`
	Date            *time.Time `json:"date"`
	Notes           *string    `json:"notes"`
}

type FuelSummary struct {
	TotalSpent   float64              `json:"total_spent"`
	TotalFills   int                  `json:"total_fills"`
	TotalLitres  float64              `json:"total_quantity"`
	ByFuelType   []FuelTypeBreakdown  `json:"by_fuel_type"`
}

type FuelTypeBreakdown struct {
	FuelType   string  `json:"fuel_type" db:"fuel_type"`
	TotalSpent float64 `json:"total_spent" db:"total_spent"`
	TotalFills int     `json:"total_fills" db:"total_fills"`
	TotalQty   float64 `json:"total_quantity" db:"total_quantity"`
}
