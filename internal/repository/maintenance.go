package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"carhistory/internal/models"
)

type MaintenanceRepository struct {
	db *sqlx.DB
}

func NewMaintenanceRepository(db *sqlx.DB) *MaintenanceRepository {
	return &MaintenanceRepository{db: db}
}

func (r *MaintenanceRepository) GetTypes(ctx context.Context) ([]models.MaintenanceType, error) {
	var types []models.MaintenanceType
	err := r.db.SelectContext(ctx, &types, `SELECT * FROM maintenance_types WHERE is_active = true ORDER BY name`)
	if err != nil {
		return nil, err
	}
	if types == nil {
		types = []models.MaintenanceType{}
	}
	return types, nil
}

func (r *MaintenanceRepository) Create(ctx context.Context, rec *models.MaintenanceRecord) error {
	query := `
		INSERT INTO maintenance_records (id, vehicle_id, maintenance_type_id, title, date_performed,
		    odometer_at_maintenance, next_due_date, next_due_odometer, cost, service_provider,
		    description, remarks, created_by, created_at, updated_at)
		VALUES (:id, :vehicle_id, :maintenance_type_id, :title, :date_performed,
		    :odometer_at_maintenance, :next_due_date, :next_due_odometer, :cost, :service_provider,
		    :description, :remarks, :created_by, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, rec)
	return err
}

func (r *MaintenanceRepository) FindByID(ctx context.Context, id, vehicleID uuid.UUID) (*models.MaintenanceRecord, error) {
	var rec models.MaintenanceRecord
	query := `
		SELECT mr.*, mt.name AS maintenance_type_name
		FROM maintenance_records mr
		JOIN maintenance_types mt ON mt.id = mr.maintenance_type_id
		WHERE mr.id = $1 AND mr.vehicle_id = $2`
	err := r.db.GetContext(ctx, &rec, query, id, vehicleID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &rec, err
}

func (r *MaintenanceRepository) FindByVehicle(ctx context.Context, vehicleID uuid.UUID, limit, offset int) ([]models.MaintenanceRecord, int64, error) {
	var records []models.MaintenanceRecord
	var total int64

	err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM maintenance_records WHERE vehicle_id = $1`, vehicleID)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT mr.*, mt.name AS maintenance_type_name
		FROM maintenance_records mr
		JOIN maintenance_types mt ON mt.id = mr.maintenance_type_id
		WHERE mr.vehicle_id = $1
		ORDER BY mr.date_performed DESC
		LIMIT $2 OFFSET $3`
	err = r.db.SelectContext(ctx, &records, query, vehicleID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if records == nil {
		records = []models.MaintenanceRecord{}
	}
	return records, total, nil
}

func (r *MaintenanceRepository) Update(ctx context.Context, rec *models.MaintenanceRecord) error {
	query := `
		UPDATE maintenance_records SET
			maintenance_type_id = :maintenance_type_id,
			title = :title,
			date_performed = :date_performed,
			odometer_at_maintenance = :odometer_at_maintenance,
			next_due_date = :next_due_date,
			next_due_odometer = :next_due_odometer,
			cost = :cost,
			service_provider = :service_provider,
			description = :description,
			remarks = :remarks,
			updated_at = :updated_at
		WHERE id = :id AND vehicle_id = :vehicle_id`
	_, err := r.db.NamedExecContext(ctx, query, rec)
	return err
}

func (r *MaintenanceRepository) Delete(ctx context.Context, id, vehicleID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM maintenance_records WHERE id = $1 AND vehicle_id = $2`, id, vehicleID)
	return err
}

// ============================================================
// SERVICE RECORDS
// ============================================================

type ServiceRepository struct {
	db *sqlx.DB
}

func NewServiceRepository(db *sqlx.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

func (r *ServiceRepository) Create(ctx context.Context, rec *models.ServiceRecord) error {
	query := `
		INSERT INTO service_records (id, vehicle_id, service_date, odometer, service_type, service_centre,
		    technician, labour_cost, parts_cost, other_cost, total_cost, next_service_date,
		    next_service_km, remarks, created_by, created_at, updated_at)
		VALUES (:id, :vehicle_id, :service_date, :odometer, :service_type, :service_centre,
		    :technician, :labour_cost, :parts_cost, :other_cost, :total_cost, :next_service_date,
		    :next_service_km, :remarks, :created_by, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, rec)
	return err
}

func (r *ServiceRepository) CreateItem(ctx context.Context, item *models.ServiceItem) error {
	query := `
		INSERT INTO service_items (id, service_record_id, name, description, cost, created_at)
		VALUES (:id, :service_record_id, :name, :description, :cost, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, item)
	return err
}

func (r *ServiceRepository) FindByID(ctx context.Context, id, vehicleID uuid.UUID) (*models.ServiceRecord, error) {
	var rec models.ServiceRecord
	err := r.db.GetContext(ctx, &rec,
		`SELECT * FROM service_records WHERE id = $1 AND vehicle_id = $2`, id, vehicleID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var items []models.ServiceItem
	err = r.db.SelectContext(ctx, &items,
		`SELECT * FROM service_items WHERE service_record_id = $1 ORDER BY created_at`, rec.ID)
	if err != nil {
		return nil, err
	}
	rec.Items = items
	return &rec, nil
}

func (r *ServiceRepository) FindByVehicle(ctx context.Context, vehicleID uuid.UUID, limit, offset int) ([]models.ServiceRecord, int64, error) {
	var records []models.ServiceRecord
	var total int64

	err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM service_records WHERE vehicle_id = $1`, vehicleID)
	if err != nil {
		return nil, 0, err
	}

	err = r.db.SelectContext(ctx, &records,
		`SELECT * FROM service_records WHERE vehicle_id = $1 ORDER BY service_date DESC LIMIT $2 OFFSET $3`,
		vehicleID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if records == nil {
		records = []models.ServiceRecord{}
	}
	return records, total, nil
}

func (r *ServiceRepository) Update(ctx context.Context, rec *models.ServiceRecord) error {
	query := `
		UPDATE service_records SET
			service_date = :service_date,
			odometer = :odometer,
			service_type = :service_type,
			service_centre = :service_centre,
			technician = :technician,
			labour_cost = :labour_cost,
			parts_cost = :parts_cost,
			other_cost = :other_cost,
			total_cost = :total_cost,
			next_service_date = :next_service_date,
			next_service_km = :next_service_km,
			remarks = :remarks,
			updated_at = :updated_at
		WHERE id = :id AND vehicle_id = :vehicle_id`
	_, err := r.db.NamedExecContext(ctx, query, rec)
	return err
}

func (r *ServiceRepository) DeleteItems(ctx context.Context, serviceRecordID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM service_items WHERE service_record_id = $1`, serviceRecordID)
	return err
}

func (r *ServiceRepository) Delete(ctx context.Context, id, vehicleID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM service_records WHERE id = $1 AND vehicle_id = $2`, id, vehicleID)
	return err
}

func (r *ServiceRepository) GetLatest(ctx context.Context, vehicleID uuid.UUID) (*models.ServiceRecord, error) {
	var rec models.ServiceRecord
	err := r.db.GetContext(ctx, &rec,
		`SELECT * FROM service_records WHERE vehicle_id = $1 ORDER BY service_date DESC LIMIT 1`, vehicleID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &rec, err
}

// ============================================================
// INSURANCE
// ============================================================

type InsuranceRepository struct {
	db *sqlx.DB
}

func NewInsuranceRepository(db *sqlx.DB) *InsuranceRepository {
	return &InsuranceRepository{db: db}
}

func (r *InsuranceRepository) Create(ctx context.Context, rec *models.InsuranceRecord) error {
	query := `
		INSERT INTO insurance_records (id, vehicle_id, insurance_company, policy_number, policy_type,
		    start_date, expiry_date, premium, coverage_amount, remarks, created_by, created_at, updated_at)
		VALUES (:id, :vehicle_id, :insurance_company, :policy_number, :policy_type,
		    :start_date, :expiry_date, :premium, :coverage_amount, :remarks, :created_by, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, rec)
	return err
}

func (r *InsuranceRepository) FindByID(ctx context.Context, id, vehicleID uuid.UUID) (*models.InsuranceRecord, error) {
	var rec models.InsuranceRecord
	err := r.db.GetContext(ctx, &rec,
		`SELECT * FROM insurance_records WHERE id = $1 AND vehicle_id = $2`, id, vehicleID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &rec, err
}

func (r *InsuranceRepository) FindByVehicle(ctx context.Context, vehicleID uuid.UUID) ([]models.InsuranceRecord, error) {
	var records []models.InsuranceRecord
	err := r.db.SelectContext(ctx, &records,
		`SELECT * FROM insurance_records WHERE vehicle_id = $1 ORDER BY expiry_date DESC`, vehicleID)
	if err != nil {
		return nil, err
	}
	if records == nil {
		records = []models.InsuranceRecord{}
	}
	return records, nil
}

func (r *InsuranceRepository) GetLatest(ctx context.Context, vehicleID uuid.UUID) (*models.InsuranceRecord, error) {
	var rec models.InsuranceRecord
	err := r.db.GetContext(ctx, &rec,
		`SELECT * FROM insurance_records WHERE vehicle_id = $1 ORDER BY expiry_date DESC LIMIT 1`, vehicleID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &rec, err
}

func (r *InsuranceRepository) Update(ctx context.Context, rec *models.InsuranceRecord) error {
	query := `
		UPDATE insurance_records SET
			insurance_company = :insurance_company,
			policy_number = :policy_number,
			policy_type = :policy_type,
			start_date = :start_date,
			expiry_date = :expiry_date,
			premium = :premium,
			coverage_amount = :coverage_amount,
			remarks = :remarks,
			updated_at = :updated_at
		WHERE id = :id AND vehicle_id = :vehicle_id`
	_, err := r.db.NamedExecContext(ctx, query, rec)
	return err
}

func (r *InsuranceRepository) Delete(ctx context.Context, id, vehicleID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM insurance_records WHERE id = $1 AND vehicle_id = $2`, id, vehicleID)
	return err
}

// ============================================================
// PUC
// ============================================================

type PUCRepository struct {
	db *sqlx.DB
}

func NewPUCRepository(db *sqlx.DB) *PUCRepository {
	return &PUCRepository{db: db}
}

func (r *PUCRepository) Create(ctx context.Context, rec *models.PUCRecord) error {
	query := `
		INSERT INTO puc_records (id, vehicle_id, certificate_number, test_date, expiry_date,
		    testing_centre, emission_info, remarks, created_by, created_at, updated_at)
		VALUES (:id, :vehicle_id, :certificate_number, :test_date, :expiry_date,
		    :testing_centre, :emission_info, :remarks, :created_by, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, rec)
	return err
}

func (r *PUCRepository) FindByID(ctx context.Context, id, vehicleID uuid.UUID) (*models.PUCRecord, error) {
	var rec models.PUCRecord
	err := r.db.GetContext(ctx, &rec,
		`SELECT * FROM puc_records WHERE id = $1 AND vehicle_id = $2`, id, vehicleID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &rec, err
}

func (r *PUCRepository) FindByVehicle(ctx context.Context, vehicleID uuid.UUID) ([]models.PUCRecord, error) {
	var records []models.PUCRecord
	err := r.db.SelectContext(ctx, &records,
		`SELECT * FROM puc_records WHERE vehicle_id = $1 ORDER BY expiry_date DESC`, vehicleID)
	if err != nil {
		return nil, err
	}
	if records == nil {
		records = []models.PUCRecord{}
	}
	return records, nil
}

func (r *PUCRepository) GetLatest(ctx context.Context, vehicleID uuid.UUID) (*models.PUCRecord, error) {
	var rec models.PUCRecord
	err := r.db.GetContext(ctx, &rec,
		`SELECT * FROM puc_records WHERE vehicle_id = $1 ORDER BY expiry_date DESC LIMIT 1`, vehicleID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &rec, err
}

func (r *PUCRepository) Update(ctx context.Context, rec *models.PUCRecord) error {
	query := `
		UPDATE puc_records SET
			certificate_number = :certificate_number,
			test_date = :test_date,
			expiry_date = :expiry_date,
			testing_centre = :testing_centre,
			emission_info = :emission_info,
			remarks = :remarks,
			updated_at = :updated_at
		WHERE id = :id AND vehicle_id = :vehicle_id`
	_, err := r.db.NamedExecContext(ctx, query, rec)
	return err
}

func (r *PUCRepository) Delete(ctx context.Context, id, vehicleID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM puc_records WHERE id = $1 AND vehicle_id = $2`, id, vehicleID)
	return err
}

// ============================================================
// TYRE
// ============================================================

type TyreRepository struct {
	db *sqlx.DB
}

func NewTyreRepository(db *sqlx.DB) *TyreRepository {
	return &TyreRepository{db: db}
}

func (r *TyreRepository) Create(ctx context.Context, rec *models.TyreRecord) error {
	query := `
		INSERT INTO tyre_records (id, vehicle_id, brand, model, size, position, purchase_date,
		    installation_date, installation_km, removal_date, removal_km, cost, warranty_expiry,
		    remarks, created_by, created_at, updated_at)
		VALUES (:id, :vehicle_id, :brand, :model, :size, :position, :purchase_date,
		    :installation_date, :installation_km, :removal_date, :removal_km, :cost, :warranty_expiry,
		    :remarks, :created_by, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, rec)
	return err
}

func (r *TyreRepository) FindByID(ctx context.Context, id, vehicleID uuid.UUID) (*models.TyreRecord, error) {
	var rec models.TyreRecord
	err := r.db.GetContext(ctx, &rec,
		`SELECT * FROM tyre_records WHERE id = $1 AND vehicle_id = $2`, id, vehicleID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &rec, err
}

func (r *TyreRepository) FindByVehicle(ctx context.Context, vehicleID uuid.UUID) ([]models.TyreRecord, error) {
	var records []models.TyreRecord
	err := r.db.SelectContext(ctx, &records,
		`SELECT * FROM tyre_records WHERE vehicle_id = $1 ORDER BY created_at DESC`, vehicleID)
	if err != nil {
		return nil, err
	}
	if records == nil {
		records = []models.TyreRecord{}
	}
	return records, nil
}

func (r *TyreRepository) Update(ctx context.Context, rec *models.TyreRecord) error {
	query := `
		UPDATE tyre_records SET
			brand = :brand, model = :model, size = :size, position = :position,
			purchase_date = :purchase_date, installation_date = :installation_date,
			installation_km = :installation_km, removal_date = :removal_date, removal_km = :removal_km,
			cost = :cost, warranty_expiry = :warranty_expiry, remarks = :remarks, updated_at = :updated_at
		WHERE id = :id AND vehicle_id = :vehicle_id`
	_, err := r.db.NamedExecContext(ctx, query, rec)
	return err
}

func (r *TyreRepository) Delete(ctx context.Context, id, vehicleID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM tyre_records WHERE id = $1 AND vehicle_id = $2`, id, vehicleID)
	return err
}

// ============================================================
// BATTERY
// ============================================================

type BatteryRepository struct {
	db *sqlx.DB
}

func NewBatteryRepository(db *sqlx.DB) *BatteryRepository {
	return &BatteryRepository{db: db}
}

func (r *BatteryRepository) Create(ctx context.Context, rec *models.BatteryRecord) error {
	query := `
		INSERT INTO battery_records (id, vehicle_id, brand, model, serial_number, installation_date,
		    installation_km, replacement_date, replacement_km, warranty_expiry, cost, remarks,
		    created_by, created_at, updated_at)
		VALUES (:id, :vehicle_id, :brand, :model, :serial_number, :installation_date,
		    :installation_km, :replacement_date, :replacement_km, :warranty_expiry, :cost, :remarks,
		    :created_by, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, rec)
	return err
}

func (r *BatteryRepository) FindByID(ctx context.Context, id, vehicleID uuid.UUID) (*models.BatteryRecord, error) {
	var rec models.BatteryRecord
	err := r.db.GetContext(ctx, &rec,
		`SELECT * FROM battery_records WHERE id = $1 AND vehicle_id = $2`, id, vehicleID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &rec, err
}

func (r *BatteryRepository) FindByVehicle(ctx context.Context, vehicleID uuid.UUID) ([]models.BatteryRecord, error) {
	var records []models.BatteryRecord
	err := r.db.SelectContext(ctx, &records,
		`SELECT * FROM battery_records WHERE vehicle_id = $1 ORDER BY created_at DESC`, vehicleID)
	if err != nil {
		return nil, err
	}
	if records == nil {
		records = []models.BatteryRecord{}
	}
	return records, nil
}

func (r *BatteryRepository) Update(ctx context.Context, rec *models.BatteryRecord) error {
	query := `
		UPDATE battery_records SET
			brand = :brand, model = :model, serial_number = :serial_number,
			installation_date = :installation_date, installation_km = :installation_km,
			replacement_date = :replacement_date, replacement_km = :replacement_km,
			warranty_expiry = :warranty_expiry, cost = :cost, remarks = :remarks, updated_at = :updated_at
		WHERE id = :id AND vehicle_id = :vehicle_id`
	_, err := r.db.NamedExecContext(ctx, query, rec)
	return err
}

func (r *BatteryRepository) Delete(ctx context.Context, id, vehicleID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM battery_records WHERE id = $1 AND vehicle_id = $2`, id, vehicleID)
	return err
}

// ============================================================
// ODOMETER
// ============================================================

type OdometerRepository struct {
	db *sqlx.DB
}

func NewOdometerRepository(db *sqlx.DB) *OdometerRepository {
	return &OdometerRepository{db: db}
}

func (r *OdometerRepository) Create(ctx context.Context, rec *models.OdometerRecord) error {
	query := `
		INSERT INTO odometer_records (id, vehicle_id, reading, recorded_at, source, notes, created_by, created_at)
		VALUES (:id, :vehicle_id, :reading, :recorded_at, :source, :notes, :created_by, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, rec)
	return err
}

func (r *OdometerRepository) FindByVehicle(ctx context.Context, vehicleID uuid.UUID, limit, offset int) ([]models.OdometerRecord, int64, error) {
	var records []models.OdometerRecord
	var total int64

	err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM odometer_records WHERE vehicle_id = $1`, vehicleID)
	if err != nil {
		return nil, 0, err
	}

	err = r.db.SelectContext(ctx, &records,
		`SELECT * FROM odometer_records WHERE vehicle_id = $1 ORDER BY recorded_at DESC LIMIT $2 OFFSET $3`,
		vehicleID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if records == nil {
		records = []models.OdometerRecord{}
	}
	return records, total, nil
}

func (r *OdometerRepository) GetLatest(ctx context.Context, vehicleID uuid.UUID) (*models.OdometerRecord, error) {
	var rec models.OdometerRecord
	err := r.db.GetContext(ctx, &rec,
		`SELECT * FROM odometer_records WHERE vehicle_id = $1 ORDER BY recorded_at DESC LIMIT 1`, vehicleID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &rec, err
}

// ============================================================
// REMINDER
// ============================================================

type ReminderRepository struct {
	db *sqlx.DB
}

func NewReminderRepository(db *sqlx.DB) *ReminderRepository {
	return &ReminderRepository{db: db}
}

func (r *ReminderRepository) Create(ctx context.Context, rem *models.Reminder) error {
	query := `
		INSERT INTO reminders (id, vehicle_id, title, reminder_type, due_date, due_odometer,
		    advance_days, advance_km, related_entity_type, related_entity_id, is_active,
		    created_by, created_at, updated_at)
		VALUES (:id, :vehicle_id, :title, :reminder_type, :due_date, :due_odometer,
		    :advance_days, :advance_km, :related_entity_type, :related_entity_id, :is_active,
		    :created_by, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, rem)
	return err
}

func (r *ReminderRepository) FindByID(ctx context.Context, id, vehicleID uuid.UUID) (*models.Reminder, error) {
	var rem models.Reminder
	err := r.db.GetContext(ctx, &rem,
		`SELECT * FROM reminders WHERE id = $1 AND vehicle_id = $2`, id, vehicleID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &rem, err
}

func (r *ReminderRepository) FindByVehicle(ctx context.Context, vehicleID uuid.UUID) ([]models.Reminder, error) {
	var reminders []models.Reminder
	err := r.db.SelectContext(ctx, &reminders,
		`SELECT * FROM reminders WHERE vehicle_id = $1 ORDER BY created_at DESC`, vehicleID)
	if err != nil {
		return nil, err
	}
	if reminders == nil {
		reminders = []models.Reminder{}
	}
	return reminders, nil
}

func (r *ReminderRepository) FindActiveReminders(ctx context.Context) ([]models.Reminder, error) {
	var reminders []models.Reminder
	err := r.db.SelectContext(ctx, &reminders,
		`SELECT * FROM reminders WHERE is_active = true`)
	if err != nil {
		return nil, err
	}
	if reminders == nil {
		reminders = []models.Reminder{}
	}
	return reminders, nil
}

func (r *ReminderRepository) Update(ctx context.Context, rem *models.Reminder) error {
	query := `
		UPDATE reminders SET
			title = :title,
			due_date = :due_date,
			due_odometer = :due_odometer,
			advance_days = :advance_days,
			advance_km = :advance_km,
			is_active = :is_active,
			updated_at = :updated_at
		WHERE id = :id AND vehicle_id = :vehicle_id`
	_, err := r.db.NamedExecContext(ctx, query, rem)
	return err
}

func (r *ReminderRepository) Delete(ctx context.Context, id, vehicleID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM reminders WHERE id = $1 AND vehicle_id = $2`, id, vehicleID)
	return err
}

// ============================================================
// NOTIFICATION
// ============================================================

type NotificationRepository struct {
	db *sqlx.DB
}

func NewNotificationRepository(db *sqlx.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, n *models.Notification) error {
	query := `
		INSERT INTO notifications (id, user_id, vehicle_id, reminder_id, channel, status,
		    subject, message, sent_at, error_message, created_at)
		VALUES (:id, :user_id, :vehicle_id, :reminder_id, :channel, :status,
		    :subject, :message, :sent_at, :error_message, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, n)
	return err
}

func (r *NotificationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, sentAt *time.Time, errMsg *string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET status = $1, sent_at = $2, error_message = $3 WHERE id = $4`,
		status, sentAt, errMsg, id)
	return err
}

func (r *NotificationRepository) FindByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1`, userID)
	if err != nil {
		return nil, 0, err
	}

	err = r.db.SelectContext(ctx, &notifications,
		`SELECT * FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if notifications == nil {
		notifications = []models.Notification{}
	}
	return notifications, total, nil
}

func (r *NotificationRepository) FindPendingForReminder(ctx context.Context, reminderID uuid.UUID, channel string) (*models.Notification, error) {
	var n models.Notification
	err := r.db.GetContext(ctx, &n,
		`SELECT * FROM notifications WHERE reminder_id = $1 AND channel = $2 AND status = 'pending'`,
		reminderID, channel)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &n, err
}

// ============================================================
// DOCUMENT
// ============================================================

type DocumentRepository struct {
	db *sqlx.DB
}

func NewDocumentRepository(db *sqlx.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) Create(ctx context.Context, doc *models.Document) error {
	query := `
		INSERT INTO documents (id, vehicle_id, related_entity_type, related_entity_id, document_type,
		    file_name, file_path, file_size, mime_type, uploaded_by, created_at)
		VALUES (:id, :vehicle_id, :related_entity_type, :related_entity_id, :document_type,
		    :file_name, :file_path, :file_size, :mime_type, :uploaded_by, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, doc)
	return err
}

func (r *DocumentRepository) FindByID(ctx context.Context, id, vehicleID uuid.UUID) (*models.Document, error) {
	var doc models.Document
	err := r.db.GetContext(ctx, &doc,
		`SELECT * FROM documents WHERE id = $1 AND vehicle_id = $2`, id, vehicleID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &doc, err
}

func (r *DocumentRepository) FindByVehicle(ctx context.Context, vehicleID uuid.UUID) ([]models.Document, error) {
	var docs []models.Document
	err := r.db.SelectContext(ctx, &docs,
		`SELECT * FROM documents WHERE vehicle_id = $1 ORDER BY created_at DESC`, vehicleID)
	if err != nil {
		return nil, err
	}
	if docs == nil {
		docs = []models.Document{}
	}
	return docs, nil
}

func (r *DocumentRepository) Delete(ctx context.Context, id, vehicleID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM documents WHERE id = $1 AND vehicle_id = $2`, id, vehicleID)
	return err
}

// ============================================================
// EXPENSE
// ============================================================

type ExpenseRepository struct {
	db *sqlx.DB
}

func NewExpenseRepository(db *sqlx.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (r *ExpenseRepository) Create(ctx context.Context, exp *models.Expense) error {
	query := `
		INSERT INTO expenses (id, vehicle_id, category, related_entity_type, related_entity_id,
		    expense_date, description, amount, vendor, odometer, notes, created_by, created_at, updated_at)
		VALUES (:id, :vehicle_id, :category, :related_entity_type, :related_entity_id,
		    :expense_date, :description, :amount, :vendor, :odometer, :notes, :created_by, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, exp)
	return err
}

func (r *ExpenseRepository) FindByID(ctx context.Context, id, vehicleID uuid.UUID) (*models.Expense, error) {
	var exp models.Expense
	err := r.db.GetContext(ctx, &exp,
		`SELECT * FROM expenses WHERE id = $1 AND vehicle_id = $2`, id, vehicleID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &exp, err
}

func (r *ExpenseRepository) FindByVehicle(ctx context.Context, vehicleID uuid.UUID, limit, offset int) ([]models.Expense, int64, error) {
	var expenses []models.Expense
	var total int64

	err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM expenses WHERE vehicle_id = $1`, vehicleID)
	if err != nil {
		return nil, 0, err
	}

	err = r.db.SelectContext(ctx, &expenses,
		`SELECT * FROM expenses WHERE vehicle_id = $1 ORDER BY expense_date DESC LIMIT $2 OFFSET $3`,
		vehicleID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if expenses == nil {
		expenses = []models.Expense{}
	}
	return expenses, total, nil
}

func (r *ExpenseRepository) Update(ctx context.Context, exp *models.Expense) error {
	query := `
		UPDATE expenses SET
			category = :category,
			expense_date = :expense_date,
			description = :description,
			amount = :amount,
			vendor = :vendor,
			odometer = :odometer,
			notes = :notes,
			updated_at = :updated_at
		WHERE id = :id AND vehicle_id = :vehicle_id`
	_, err := r.db.NamedExecContext(ctx, query, exp)
	return err
}

func (r *ExpenseRepository) Delete(ctx context.Context, id, vehicleID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM expenses WHERE id = $1 AND vehicle_id = $2`, id, vehicleID)
	return err
}

func (r *ExpenseRepository) GetSummary(ctx context.Context, vehicleID uuid.UUID) (map[string]float64, error) {
	rows, err := r.db.QueryxContext(ctx,
		`SELECT category, COALESCE(SUM(amount), 0) AS total FROM expenses WHERE vehicle_id = $1 GROUP BY category`,
		vehicleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var category string
		var total float64
		if err := rows.Scan(&category, &total); err != nil {
			return nil, err
		}
		result[category] = total
	}
	return result, nil
}

func (r *ExpenseRepository) GetTotalByPeriod(ctx context.Context, vehicleID uuid.UUID, startDate, endDate time.Time) (float64, error) {
	var total float64
	err := r.db.GetContext(ctx, &total,
		`SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE vehicle_id = $1 AND expense_date BETWEEN $2 AND $3`,
		vehicleID, startDate, endDate)
	return total, err
}

// ============================================================
// ACTIVITY LOG
// ============================================================

type ActivityRepository struct {
	db *sqlx.DB
}

func NewActivityRepository(db *sqlx.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

func (r *ActivityRepository) Create(ctx context.Context, log *models.ActivityLog) error {
	query := `
		INSERT INTO activity_logs (id, vehicle_id, user_id, action, entity, entity_id, description, created_at)
		VALUES (:id, :vehicle_id, :user_id, :action, :entity, :entity_id, :description, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, log)
	return err
}

func (r *ActivityRepository) FindByVehicle(ctx context.Context, vehicleID uuid.UUID, limit, offset int) ([]models.ActivityLog, int64, error) {
	var logs []models.ActivityLog
	var total int64

	err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM activity_logs WHERE vehicle_id = $1`, vehicleID)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT al.*, u.name AS user_name
		FROM activity_logs al
		JOIN users u ON u.id = al.user_id
		WHERE al.vehicle_id = $1
		ORDER BY al.created_at DESC
		LIMIT $2 OFFSET $3`
	err = r.db.SelectContext(ctx, &logs, query, vehicleID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if logs == nil {
		logs = []models.ActivityLog{}
	}
	return logs, total, nil
}

func (r *ActivityRepository) FindRecentByVehicle(ctx context.Context, vehicleID uuid.UUID, n int) ([]models.ActivityLog, error) {
	var logs []models.ActivityLog
	query := `
		SELECT al.*, u.name AS user_name
		FROM activity_logs al
		JOIN users u ON u.id = al.user_id
		WHERE al.vehicle_id = $1
		ORDER BY al.created_at DESC
		LIMIT $2`
	err := r.db.SelectContext(ctx, &logs, query, vehicleID, n)
	if err != nil {
		return nil, err
	}
	if logs == nil {
		logs = []models.ActivityLog{}
	}
	return logs, nil
}
