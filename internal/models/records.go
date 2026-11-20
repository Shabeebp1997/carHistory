package models

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// INSURANCE
// ============================================================

type InsuranceRecord struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	VehicleID        uuid.UUID  `json:"vehicle_id" db:"vehicle_id"`
	InsuranceCompany string     `json:"insurance_company" db:"insurance_company"`
	PolicyNumber     *string    `json:"policy_number" db:"policy_number"`
	PolicyType       *string    `json:"policy_type" db:"policy_type"`
	StartDate        time.Time  `json:"start_date" db:"start_date"`
	ExpiryDate       time.Time  `json:"expiry_date" db:"expiry_date"`
	Premium          *float64   `json:"premium" db:"premium"`
	CoverageAmount   *float64   `json:"coverage_amount" db:"coverage_amount"`
	Remarks          *string    `json:"remarks" db:"remarks"`
	Status           string     `json:"status,omitempty"`
	CreatedBy        uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateInsuranceRequest struct {
	InsuranceCompany string    `json:"insurance_company" validate:"required"`
	PolicyNumber     *string   `json:"policy_number"`
	PolicyType       *string   `json:"policy_type"`
	StartDate        time.Time `json:"start_date" validate:"required"`
	ExpiryDate       time.Time `json:"expiry_date" validate:"required"`
	Premium          *float64  `json:"premium"`
	CoverageAmount   *float64  `json:"coverage_amount"`
	Remarks          *string   `json:"remarks"`
}

type UpdateInsuranceRequest struct {
	InsuranceCompany *string    `json:"insurance_company"`
	PolicyNumber     *string    `json:"policy_number"`
	PolicyType       *string    `json:"policy_type"`
	StartDate        *time.Time `json:"start_date"`
	ExpiryDate       *time.Time `json:"expiry_date"`
	Premium          *float64   `json:"premium"`
	CoverageAmount   *float64   `json:"coverage_amount"`
	Remarks          *string    `json:"remarks"`
}

// ============================================================
// PUC
// ============================================================

type PUCRecord struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	VehicleID         uuid.UUID  `json:"vehicle_id" db:"vehicle_id"`
	CertificateNumber *string    `json:"certificate_number" db:"certificate_number"`
	TestDate          time.Time  `json:"test_date" db:"test_date"`
	ExpiryDate        time.Time  `json:"expiry_date" db:"expiry_date"`
	TestingCentre     *string    `json:"testing_centre" db:"testing_centre"`
	EmissionInfo      *string    `json:"emission_info" db:"emission_info"`
	Remarks           *string    `json:"remarks" db:"remarks"`
	Status            string     `json:"status,omitempty"`
	CreatedBy         uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

type CreatePUCRequest struct {
	CertificateNumber *string   `json:"certificate_number"`
	TestDate          time.Time `json:"test_date" validate:"required"`
	ExpiryDate        time.Time `json:"expiry_date" validate:"required"`
	TestingCentre     *string   `json:"testing_centre"`
	EmissionInfo      *string   `json:"emission_info"`
	Remarks           *string   `json:"remarks"`
}

type UpdatePUCRequest struct {
	CertificateNumber *string    `json:"certificate_number"`
	TestDate          *time.Time `json:"test_date"`
	ExpiryDate        *time.Time `json:"expiry_date"`
	TestingCentre     *string    `json:"testing_centre"`
	EmissionInfo      *string    `json:"emission_info"`
	Remarks           *string    `json:"remarks"`
}

// ============================================================
// TYRE
// ============================================================

type TyreRecord struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	VehicleID        uuid.UUID  `json:"vehicle_id" db:"vehicle_id"`
	Brand            *string    `json:"brand" db:"brand"`
	Model            *string    `json:"model" db:"model"`
	Size             *string    `json:"size" db:"size"`
	Position         *string    `json:"position" db:"position"`
	PurchaseDate     *time.Time `json:"purchase_date" db:"purchase_date"`
	InstallationDate *time.Time `json:"installation_date" db:"installation_date"`
	InstallationKM   *float64   `json:"installation_km" db:"installation_km"`
	RemovalDate      *time.Time `json:"removal_date" db:"removal_date"`
	RemovalKM        *float64   `json:"removal_km" db:"removal_km"`
	Cost             *float64   `json:"cost" db:"cost"`
	WarrantyExpiry   *time.Time `json:"warranty_expiry" db:"warranty_expiry"`
	Remarks          *string    `json:"remarks" db:"remarks"`
	CreatedBy        uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateTyreRequest struct {
	Brand            *string    `json:"brand"`
	Model            *string    `json:"model"`
	Size             *string    `json:"size"`
	Position         *string    `json:"position"`
	PurchaseDate     *time.Time `json:"purchase_date"`
	InstallationDate *time.Time `json:"installation_date"`
	InstallationKM   *float64   `json:"installation_km"`
	RemovalDate      *time.Time `json:"removal_date"`
	RemovalKM        *float64   `json:"removal_km"`
	Cost             *float64   `json:"cost"`
	WarrantyExpiry   *time.Time `json:"warranty_expiry"`
	Remarks          *string    `json:"remarks"`
}

type UpdateTyreRequest = CreateTyreRequest

// ============================================================
// BATTERY
// ============================================================

type BatteryRecord struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	VehicleID        uuid.UUID  `json:"vehicle_id" db:"vehicle_id"`
	Brand            *string    `json:"brand" db:"brand"`
	Model            *string    `json:"model" db:"model"`
	SerialNumber     *string    `json:"serial_number" db:"serial_number"`
	InstallationDate *time.Time `json:"installation_date" db:"installation_date"`
	InstallationKM   *float64   `json:"installation_km" db:"installation_km"`
	ReplacementDate  *time.Time `json:"replacement_date" db:"replacement_date"`
	ReplacementKM    *float64   `json:"replacement_km" db:"replacement_km"`
	WarrantyExpiry   *time.Time `json:"warranty_expiry" db:"warranty_expiry"`
	Cost             *float64   `json:"cost" db:"cost"`
	Remarks          *string    `json:"remarks" db:"remarks"`
	CreatedBy        uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateBatteryRequest struct {
	Brand            *string    `json:"brand"`
	Model            *string    `json:"model"`
	SerialNumber     *string    `json:"serial_number"`
	InstallationDate *time.Time `json:"installation_date"`
	InstallationKM   *float64   `json:"installation_km"`
	ReplacementDate  *time.Time `json:"replacement_date"`
	ReplacementKM    *float64   `json:"replacement_km"`
	WarrantyExpiry   *time.Time `json:"warranty_expiry"`
	Cost             *float64   `json:"cost"`
	Remarks          *string    `json:"remarks"`
}

type UpdateBatteryRequest = CreateBatteryRequest

// ============================================================
// ODOMETER
// ============================================================

type OdometerRecord struct {
	ID         uuid.UUID `json:"id" db:"id"`
	VehicleID  uuid.UUID `json:"vehicle_id" db:"vehicle_id"`
	Reading    float64   `json:"reading" db:"reading"`
	RecordedAt time.Time `json:"recorded_at" db:"recorded_at"`
	Source     *string   `json:"source" db:"source"`
	Notes      *string   `json:"notes" db:"notes"`
	CreatedBy  uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type CreateOdometerRequest struct {
	Reading    float64    `json:"reading" validate:"required,gt=0"`
	RecordedAt *time.Time `json:"recorded_at"`
	Source     *string    `json:"source"`
	Notes      *string    `json:"notes"`
}

// ============================================================
// REMINDER
// ============================================================

type Reminder struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	VehicleID         uuid.UUID  `json:"vehicle_id" db:"vehicle_id"`
	Title             string     `json:"title" db:"title"`
	ReminderType      string     `json:"reminder_type" db:"reminder_type"`
	DueDate           *time.Time `json:"due_date" db:"due_date"`
	DueOdometer       *float64   `json:"due_odometer" db:"due_odometer"`
	AdvanceDays       *int       `json:"advance_days" db:"advance_days"`
	AdvanceKM         *float64   `json:"advance_km" db:"advance_km"`
	RelatedEntityType *string    `json:"related_entity_type" db:"related_entity_type"`
	RelatedEntityID   *uuid.UUID `json:"related_entity_id" db:"related_entity_id"`
	IsActive          bool       `json:"is_active" db:"is_active"`
	CreatedBy         uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateReminderRequest struct {
	Title             string     `json:"title" validate:"required,max=255"`
	ReminderType      string     `json:"reminder_type" validate:"required,oneof=date odometer date_or_odometer"`
	DueDate           *time.Time `json:"due_date"`
	DueOdometer       *float64   `json:"due_odometer"`
	AdvanceDays       *int       `json:"advance_days"`
	AdvanceKM         *float64   `json:"advance_km"`
	RelatedEntityType *string    `json:"related_entity_type"`
	RelatedEntityID   *uuid.UUID `json:"related_entity_id"`
}

type UpdateReminderRequest struct {
	Title             *string    `json:"title"`
	DueDate           *time.Time `json:"due_date"`
	DueOdometer       *float64   `json:"due_odometer"`
	AdvanceDays       *int       `json:"advance_days"`
	AdvanceKM         *float64   `json:"advance_km"`
	IsActive          *bool      `json:"is_active"`
}

// ============================================================
// NOTIFICATION
// ============================================================

type Notification struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	VehicleID    *uuid.UUID `json:"vehicle_id" db:"vehicle_id"`
	ReminderID   *uuid.UUID `json:"reminder_id" db:"reminder_id"`
	Channel      string     `json:"channel" db:"channel"`
	Status       string     `json:"status" db:"status"`
	Subject      *string    `json:"subject" db:"subject"`
	Message      string     `json:"message" db:"message"`
	SentAt       *time.Time `json:"sent_at" db:"sent_at"`
	ErrorMessage *string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// ============================================================
// DOCUMENT
// ============================================================

type Document struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	VehicleID         uuid.UUID  `json:"vehicle_id" db:"vehicle_id"`
	RelatedEntityType *string    `json:"related_entity_type" db:"related_entity_type"`
	RelatedEntityID   *uuid.UUID `json:"related_entity_id" db:"related_entity_id"`
	DocumentType      *string    `json:"document_type" db:"document_type"`
	FileName          string     `json:"file_name" db:"file_name"`
	FilePath          string     `json:"file_path" db:"file_path"`
	FileURL           string     `json:"file_url,omitempty"`
	FileSize          *int64     `json:"file_size" db:"file_size"`
	MimeType          *string    `json:"mime_type" db:"mime_type"`
	UploadedBy        uuid.UUID  `json:"uploaded_by" db:"uploaded_by"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

// ============================================================
// EXPENSE
// ============================================================

type Expense struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	VehicleID         uuid.UUID  `json:"vehicle_id" db:"vehicle_id"`
	Category          string     `json:"category" db:"category"`
	RelatedEntityType *string    `json:"related_entity_type" db:"related_entity_type"`
	RelatedEntityID   *uuid.UUID `json:"related_entity_id" db:"related_entity_id"`
	ExpenseDate       time.Time  `json:"expense_date" db:"expense_date"`
	Description       *string    `json:"description" db:"description"`
	Amount            float64    `json:"amount" db:"amount"`
	Vendor            *string    `json:"vendor" db:"vendor"`
	Odometer          *float64   `json:"odometer" db:"odometer"`
	Notes             *string    `json:"notes" db:"notes"`
	CreatedBy         uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateExpenseRequest struct {
	Category    string     `json:"category" validate:"required,oneof=service maintenance insurance tyre battery fuel repair other"`
	ExpenseDate time.Time  `json:"expense_date" validate:"required"`
	Description *string    `json:"description"`
	Amount      float64    `json:"amount" validate:"required,gt=0"`
	Vendor      *string    `json:"vendor"`
	Odometer    *float64   `json:"odometer"`
	Notes       *string    `json:"notes"`
}

type UpdateExpenseRequest struct {
	Category    *string    `json:"category"`
	ExpenseDate *time.Time `json:"expense_date"`
	Description *string    `json:"description"`
	Amount      *float64   `json:"amount"`
	Vendor      *string    `json:"vendor"`
	Odometer    *float64   `json:"odometer"`
	Notes       *string    `json:"notes"`
}

// ============================================================
// ACTIVITY LOG
// ============================================================

type ActivityLog struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	VehicleID   *uuid.UUID `json:"vehicle_id" db:"vehicle_id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	UserName    string     `json:"user_name,omitempty" db:"user_name"`
	Action      string     `json:"action" db:"action"`
	Entity      *string    `json:"entity" db:"entity"`
	EntityID    *uuid.UUID `json:"entity_id" db:"entity_id"`
	Description *string    `json:"description" db:"description"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// ============================================================
// DASHBOARD / REPORTS
// ============================================================

type VehicleDashboard struct {
	Vehicle            *Vehicle           `json:"vehicle"`
	CurrentOdometer    float64            `json:"current_odometer"`
	LatestInsurance    *InsuranceRecord   `json:"latest_insurance"`
	LatestPUC         *PUCRecord         `json:"latest_puc"`
	UpcomingReminders  []ReminderStatus   `json:"upcoming_reminders"`
	RecentActivity     []ActivityLog      `json:"recent_activity"`
}

type ReminderStatus struct {
	Reminder     *Reminder `json:"reminder"`
	Status       string    `json:"status"`
	DaysLeft     *int      `json:"days_left,omitempty"`
	KMLeft       *float64  `json:"km_left,omitempty"`
}

type UserDashboard struct {
	TotalVehicles     int                  `json:"total_vehicles"`
	Vehicles          []Vehicle            `json:"vehicles"`
	UpcomingReminders []ReminderStatus     `json:"upcoming_reminders"`
}

type ExpenseReport struct {
	TotalExpense         float64            `json:"total_expense"`
	AnnualExpense        float64            `json:"annual_expense"`
	MonthlyExpense       float64            `json:"monthly_expense"`
	ExpenseByCategory    map[string]float64 `json:"expense_by_category"`
	CostPerKM            *float64           `json:"cost_per_km,omitempty"`
}
