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

type VehicleRepository struct {
	db *sqlx.DB
}

func NewVehicleRepository(db *sqlx.DB) *VehicleRepository {
	return &VehicleRepository{db: db}
}

func (r *VehicleRepository) Create(ctx context.Context, v *models.Vehicle) error {
	query := `
		INSERT INTO vehicles (id, vehicle_type_id, brand, model, variant, manufacturing_year,
		    purchase_date, registration_number, chassis_number, engine_number, rc_owner,
		    owner_contact, fuel_type, current_odometer, colour, nickname, notes,
		    is_archived, created_by, created_at, updated_at)
		VALUES (:id, :vehicle_type_id, :brand, :model, :variant, :manufacturing_year,
		    :purchase_date, :registration_number, :chassis_number, :engine_number, :rc_owner,
		    :owner_contact, :fuel_type, :current_odometer, :colour, :nickname, :notes,
		    :is_archived, :created_by, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, v)
	return err
}

func (r *VehicleRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Vehicle, error) {
	var v models.Vehicle
	query := `
		SELECT v.*, vt.name AS vehicle_type_name
		FROM vehicles v
		JOIN vehicle_types vt ON vt.id = v.vehicle_type_id
		WHERE v.id = $1 AND v.is_archived = false`
	err := r.db.GetContext(ctx, &v, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &v, err
}

func (r *VehicleRepository) FindByIDWithRole(ctx context.Context, vehicleID, userID uuid.UUID) (*models.Vehicle, error) {
	var v models.Vehicle
	query := `
		SELECT v.*, vt.name AS vehicle_type_name, vm.role AS user_role
		FROM vehicles v
		JOIN vehicle_types vt ON vt.id = v.vehicle_type_id
		JOIN vehicle_members vm ON vm.vehicle_id = v.id AND vm.user_id = $2
		WHERE v.id = $1 AND v.is_archived = false`
	err := r.db.GetContext(ctx, &v, query, vehicleID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &v, err
}

func (r *VehicleRepository) FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]models.Vehicle, error) {
	var vehicles []models.Vehicle
	query := `
		SELECT v.*, vt.name AS vehicle_type_name, vm.role AS user_role
		FROM vehicles v
		JOIN vehicle_types vt ON vt.id = v.vehicle_type_id
		JOIN vehicle_members vm ON vm.vehicle_id = v.id AND vm.user_id = $1
		WHERE v.is_archived = false
		ORDER BY v.created_at DESC`
	err := r.db.SelectContext(ctx, &vehicles, query, userID)
	if err != nil {
		return nil, err
	}
	if vehicles == nil {
		vehicles = []models.Vehicle{}
	}
	return vehicles, nil
}

func (r *VehicleRepository) Update(ctx context.Context, v *models.Vehicle) error {
	query := `
		UPDATE vehicles SET
			vehicle_type_id = :vehicle_type_id,
			brand = :brand,
			model = :model,
			variant = :variant,
			manufacturing_year = :manufacturing_year,
			purchase_date = :purchase_date,
			registration_number = :registration_number,
			chassis_number = :chassis_number,
			engine_number = :engine_number,
			rc_owner = :rc_owner,
			owner_contact = :owner_contact,
			fuel_type = :fuel_type,
			current_odometer = :current_odometer,
			colour = :colour,
			nickname = :nickname,
			purchase_price = :purchase_price,
			notes = :notes,
			updated_at = :updated_at
		WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, query, v)
	return err
}

func (r *VehicleRepository) Search(ctx context.Context, registration, email, phone string) ([]models.Vehicle, error) {
	var vehicles []models.Vehicle
	query := `
		SELECT DISTINCT v.*, vt.name AS vehicle_type_name
		FROM vehicles v
		JOIN vehicle_types vt ON vt.id = v.vehicle_type_id
		JOIN vehicle_members vm ON vm.vehicle_id = v.id
		JOIN users u ON u.id = vm.user_id
		WHERE v.is_archived = false
		  AND (
		        ($1 <> '' AND v.registration_number ILIKE '%' || $1 || '%')
		     OR ($2 <> '' AND u.email = $2 AND vm.role = 'owner')
		     OR ($3 <> '' AND u.mobile_number = $3 AND vm.role = 'owner')
		  )
		ORDER BY v.created_at DESC`
	err := r.db.SelectContext(ctx, &vehicles, query, registration, email, phone)
	if vehicles == nil {
		vehicles = []models.Vehicle{}
	}
	return vehicles, err
}

func (r *VehicleRepository) Archive(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE vehicles SET is_archived = true, updated_at = $1 WHERE id = $2`,
		time.Now(), id)
	return err
}

func (r *VehicleRepository) UpdateOdometer(ctx context.Context, vehicleID uuid.UUID, reading float64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE vehicles SET current_odometer = $1, updated_at = $2 WHERE id = $3`,
		reading, time.Now(), vehicleID)
	return err
}

func (r *VehicleRepository) GetAllVehicleTypes(ctx context.Context) ([]models.VehicleType, error) {
	var types []models.VehicleType
	err := r.db.SelectContext(ctx, &types, `SELECT * FROM vehicle_types WHERE is_active = true ORDER BY name`)
	if err != nil {
		return nil, err
	}
	if types == nil {
		types = []models.VehicleType{}
	}
	return types, nil
}

func (r *VehicleRepository) AddMember(ctx context.Context, member *models.VehicleMember) error {
	query := `
		INSERT INTO vehicle_members (id, vehicle_id, user_id, role, joined_at, created_at)
		VALUES (:id, :vehicle_id, :user_id, :role, :joined_at, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, member)
	return err
}

func (r *VehicleRepository) GetMembers(ctx context.Context, vehicleID uuid.UUID) ([]models.VehicleMember, error) {
	var members []models.VehicleMember
	query := `
		SELECT vm.*, u.name AS user_name, u.email AS user_email
		FROM vehicle_members vm
		JOIN users u ON u.id = vm.user_id
		WHERE vm.vehicle_id = $1
		ORDER BY vm.role, vm.joined_at`
	err := r.db.SelectContext(ctx, &members, query, vehicleID)
	if err != nil {
		return nil, err
	}
	if members == nil {
		members = []models.VehicleMember{}
	}
	return members, nil
}

func (r *VehicleRepository) GetMember(ctx context.Context, vehicleID, userID uuid.UUID) (*models.VehicleMember, error) {
	var member models.VehicleMember
	err := r.db.GetContext(ctx, &member,
		`SELECT * FROM vehicle_members WHERE vehicle_id = $1 AND user_id = $2`,
		vehicleID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &member, err
}

func (r *VehicleRepository) UpdateMemberRole(ctx context.Context, vehicleID, userID uuid.UUID, role string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE vehicle_members SET role = $1 WHERE vehicle_id = $2 AND user_id = $3`,
		role, vehicleID, userID)
	return err
}

func (r *VehicleRepository) RemoveMember(ctx context.Context, vehicleID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM vehicle_members WHERE vehicle_id = $1 AND user_id = $2`,
		vehicleID, userID)
	return err
}

func (r *VehicleRepository) CreateInvitation(ctx context.Context, inv *models.VehicleInvitation) error {
	query := `
		INSERT INTO vehicle_invitations (id, vehicle_id, invited_by, invited_email, role, token, status, expires_at, created_at)
		VALUES (:id, :vehicle_id, :invited_by, :invited_email, :role, :token, :status, :expires_at, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, inv)
	return err
}

func (r *VehicleRepository) FindInvitationByToken(ctx context.Context, token string) (*models.VehicleInvitation, error) {
	var inv models.VehicleInvitation
	query := `
		SELECT vi.*, v.nickname AS vehicle_name, u.name AS inviter_name
		FROM vehicle_invitations vi
		JOIN vehicles v ON v.id = vi.vehicle_id
		JOIN users u ON u.id = vi.invited_by
		WHERE vi.token = $1`
	err := r.db.GetContext(ctx, &inv, query, token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &inv, err
}

func (r *VehicleRepository) UpdateInvitationStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE vehicle_invitations SET status = $1 WHERE id = $2`,
		status, id)
	return err
}

func (r *VehicleRepository) FindPendingInvitationsByEmail(ctx context.Context, email string) ([]models.VehicleInvitation, error) {
	var invitations []models.VehicleInvitation
	query := `
		SELECT vi.*, v.nickname AS vehicle_name, u.name AS inviter_name
		FROM vehicle_invitations vi
		JOIN vehicles v ON v.id = vi.vehicle_id
		JOIN users u ON u.id = vi.invited_by
		WHERE vi.invited_email = $1 AND vi.status = 'pending' AND vi.expires_at > NOW()
		ORDER BY vi.created_at DESC`
	err := r.db.SelectContext(ctx, &invitations, query, email)
	if err != nil {
		return nil, err
	}
	if invitations == nil {
		invitations = []models.VehicleInvitation{}
	}
	return invitations, nil
}

func (r *VehicleRepository) FindInvitationsByVehicle(ctx context.Context, vehicleID uuid.UUID) ([]models.VehicleInvitation, error) {
	var invitations []models.VehicleInvitation
	query := `
		SELECT vi.*, v.nickname AS vehicle_name, u.name AS inviter_name
		FROM vehicle_invitations vi
		JOIN vehicles v ON v.id = vi.vehicle_id
		JOIN users u ON u.id = vi.invited_by
		WHERE vi.vehicle_id = $1
		ORDER BY vi.created_at DESC`
	err := r.db.SelectContext(ctx, &invitations, query, vehicleID)
	if err != nil {
		return nil, err
	}
	if invitations == nil {
		invitations = []models.VehicleInvitation{}
	}
	return invitations, nil
}

func (r *VehicleRepository) ExpireOldInvitations(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE vehicle_invitations SET status = 'expired' WHERE status = 'pending' AND expires_at <= NOW()`)
	return err
}
