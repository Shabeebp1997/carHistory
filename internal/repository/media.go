package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"carhistory/internal/models"
)

type MediaRepository struct {
	db *sqlx.DB
}

func NewMediaRepository(db *sqlx.DB) *MediaRepository {
	return &MediaRepository{db: db}
}

func (r *MediaRepository) Create(ctx context.Context, m *models.Media) error {
	query := `
		INSERT INTO media (id, entity_type, entity_id, file_path, file_name, media_type, mime_type, file_size, uploaded_by, created_at)
		VALUES (:id, :entity_type, :entity_id, :file_path, :file_name, :media_type, :mime_type, :file_size, :uploaded_by, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, m)
	return err
}

func (r *MediaRepository) FindByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]models.Media, error) {
	var items []models.Media
	err := r.db.SelectContext(ctx, &items,
		`SELECT * FROM media WHERE entity_type = $1 AND entity_id = $2 ORDER BY created_at ASC`,
		entityType, entityID)
	return items, err
}

func (r *MediaRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Media, error) {
	var m models.Media
	err := r.db.GetContext(ctx, &m, `SELECT * FROM media WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &m, err
}

func (r *MediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM media WHERE id = $1`, id)
	return err
}

func (r *MediaRepository) CountAll(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM media`)
	return count, err
}
