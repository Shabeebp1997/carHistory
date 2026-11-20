package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"carhistory/internal/models"
)

type FuelRepository struct {
	db *sqlx.DB
}

func NewFuelRepository(db *sqlx.DB) *FuelRepository {
	return &FuelRepository{db: db}
}

func (r *FuelRepository) Create(ctx context.Context, f *models.FuelLog) error {
	query := `
		INSERT INTO fuel_logs
			(id, vehicle_id, logged_by, fuel_type, quantity, price_per_unit, total_cost,
			 odometer_reading, fuel_station, date, notes, created_at, updated_at)
		VALUES
			(:id, :vehicle_id, :logged_by, :fuel_type, :quantity, :price_per_unit, :total_cost,
			 :odometer_reading, :fuel_station, :date, :notes, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, f)
	return err
}

func (r *FuelRepository) List(ctx context.Context, vehicleID uuid.UUID) ([]models.FuelLog, error) {
	var logs []models.FuelLog
	err := r.db.SelectContext(ctx, &logs,
		`SELECT * FROM fuel_logs WHERE vehicle_id = $1 ORDER BY date DESC, created_at DESC`,
		vehicleID)
	return logs, err
}

func (r *FuelRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.FuelLog, error) {
	var f models.FuelLog
	err := r.db.GetContext(ctx, &f, `SELECT * FROM fuel_logs WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &f, err
}

func (r *FuelRepository) Update(ctx context.Context, f *models.FuelLog) error {
	query := `
		UPDATE fuel_logs SET
			fuel_type        = :fuel_type,
			quantity         = :quantity,
			price_per_unit   = :price_per_unit,
			total_cost       = :total_cost,
			odometer_reading = :odometer_reading,
			fuel_station     = :fuel_station,
			date             = :date,
			notes            = :notes,
			updated_at       = :updated_at
		WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, query, f)
	return err
}

func (r *FuelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM fuel_logs WHERE id = $1`, id)
	return err
}

func (r *FuelRepository) Summary(ctx context.Context, vehicleID uuid.UUID) (*models.FuelSummary, error) {
	var summary models.FuelSummary

	row := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(total_cost),0), COUNT(*), COALESCE(SUM(quantity),0)
		 FROM fuel_logs WHERE vehicle_id = $1`, vehicleID)
	if err := row.Scan(&summary.TotalSpent, &summary.TotalFills, &summary.TotalLitres); err != nil {
		return nil, err
	}

	var breakdown []models.FuelTypeBreakdown
	err := r.db.SelectContext(ctx, &breakdown,
		`SELECT fuel_type,
		        COALESCE(SUM(total_cost),0) AS total_spent,
		        COUNT(*) AS total_fills,
		        COALESCE(SUM(quantity),0) AS total_quantity
		 FROM fuel_logs WHERE vehicle_id = $1
		 GROUP BY fuel_type ORDER BY total_spent DESC`, vehicleID)
	if err != nil {
		return nil, err
	}
	summary.ByFuelType = breakdown
	return &summary, nil
}

func (r *FuelRepository) TotalCostByVehicle(ctx context.Context, vehicleID uuid.UUID, from, to *time.Time) (float64, error) {
	query := `SELECT COALESCE(SUM(total_cost),0) FROM fuel_logs WHERE vehicle_id = $1`
	args := []interface{}{vehicleID}
	n := 2
	if from != nil {
		args = append(args, *from)
		query += fmt.Sprintf(` AND date >= $%d`, n)
		n++
	}
	if to != nil {
		args = append(args, *to)
		query += fmt.Sprintf(` AND date <= $%d`, n)
	}
	var total float64
	err := r.db.GetContext(ctx, &total, query, args...)
	return total, err
}
