package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	"carhistory/internal/models"
)

type AdminRepository struct {
	db *sqlx.DB
}

func NewAdminRepository(db *sqlx.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) ListUsers(ctx context.Context, page, limit int) ([]models.AdminUserListItem, int, error) {
	offset := (page - 1) * limit
	var total int
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM users`); err != nil {
		return nil, 0, err
	}

	var users []models.AdminUserListItem
	err := r.db.SelectContext(ctx, &users, `
		SELECT u.id, u.name, u.email, u.mobile_number, u.profile_picture,
		       u.is_admin, u.is_active, u.created_at,
		       COUNT(vm.vehicle_id) AS vehicle_count
		FROM users u
		LEFT JOIN vehicle_members vm ON vm.user_id = u.id
		GROUP BY u.id
		ORDER BY u.created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	return users, total, err
}

func (r *AdminRepository) UserAnalytics(ctx context.Context) (*models.UserAnalytics, error) {
	var a models.UserAnalytics
	err := r.db.QueryRowContext(ctx, `
		SELECT
		    COUNT(*),
		    SUM(CASE WHEN is_active THEN 1 ELSE 0 END),
		    SUM(CASE WHEN is_admin  THEN 1 ELSE 0 END),
		    SUM(CASE WHEN email IS NOT NULL THEN 1 ELSE 0 END),
		    SUM(CASE WHEN mobile_number IS NOT NULL THEN 1 ELSE 0 END),
		    SUM(CASE WHEN created_at >= date_trunc('month', NOW()) THEN 1 ELSE 0 END)
		FROM users`).Scan(&a.Total, &a.Active, &a.Admins, &a.WithEmail, &a.WithPhone, &a.NewThisMonth)
	return &a, err
}

func (r *AdminRepository) VehicleAnalytics(ctx context.Context) (*models.VehicleAnalytics, error) {
	var a models.VehicleAnalytics

	row := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*), SUM(CASE WHEN NOT is_archived THEN 1 ELSE 0 END) FROM vehicles`)
	if err := row.Scan(&a.Total, &a.Active); err != nil {
		return nil, err
	}

	if err := r.db.SelectContext(ctx, &a.ByType, `
		SELECT vt.name AS type_name, COUNT(v.id) AS count
		FROM vehicles v JOIN vehicle_types vt ON v.vehicle_type_id = vt.id
		GROUP BY vt.name ORDER BY count DESC`); err != nil {
		return nil, err
	}

	if err := r.db.SelectContext(ctx, &a.ByFuelType, `
		SELECT COALESCE(fuel_type, 'unknown') AS fuel_type, COUNT(*) AS count
		FROM vehicles GROUP BY fuel_type ORDER BY count DESC`); err != nil {
		return nil, err
	}

	if err := r.db.SelectContext(ctx, &a.PopularBrands, `
		SELECT brand, COUNT(*) AS count
		FROM vehicles WHERE is_archived = false
		GROUP BY brand ORDER BY count DESC LIMIT 10`); err != nil {
		return nil, err
	}

	if err := r.db.SelectContext(ctx, &a.PopularModels, `
		SELECT brand, model, COUNT(*) AS count
		FROM vehicles WHERE is_archived = false
		GROUP BY brand, model ORDER BY count DESC LIMIT 10`); err != nil {
		return nil, err
	}

	return &a, nil
}

func (r *AdminRepository) ActivityAnalytics(ctx context.Context) (*models.ActivityAnalytics, error) {
	var a models.ActivityAnalytics
	queries := []struct {
		dest  *int
		query string
	}{
		{&a.TotalMaintenanceRecords, `SELECT COUNT(*) FROM maintenance_records`},
		{&a.TotalServiceRecords, `SELECT COUNT(*) FROM service_records`},
		{&a.TotalInsuranceRecords, `SELECT COUNT(*) FROM insurance_records`},
		{&a.TotalFuelLogs, `SELECT COUNT(*) FROM fuel_logs`},
		{&a.TotalDocuments, `SELECT COUNT(*) FROM documents`},
		{&a.TotalReminders, `SELECT COUNT(*) FROM reminders`},
		{&a.ActiveReminders, `SELECT COUNT(*) FROM reminders WHERE is_active = TRUE`},
		{&a.TotalMedia, `SELECT COUNT(*) FROM media`},
	}
	for _, q := range queries {
		if err := r.db.GetContext(ctx, q.dest, q.query); err != nil {
			return nil, err
		}
	}
	return &a, nil
}

func (r *AdminRepository) ListAdmins(ctx context.Context) ([]models.AdminUserListItem, error) {
	var admins []models.AdminUserListItem
	err := r.db.SelectContext(ctx, &admins, `
		SELECT u.id, u.name, u.email, u.mobile_number, u.profile_picture,
		       u.is_admin, u.is_active, u.created_at,
		       COUNT(vm.vehicle_id) AS vehicle_count
		FROM users u
		LEFT JOIN vehicle_members vm ON vm.user_id = u.id
		WHERE u.is_admin = TRUE
		GROUP BY u.id
		ORDER BY u.created_at ASC`)
	return admins, err
}
