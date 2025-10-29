package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/ilam072/image-processor/intenal/image/repo"
	"github.com/ilam072/image-processor/intenal/types/domain"
	"github.com/ilam072/image-processor/pkg/errutils"
	"github.com/wb-go/wbf/dbpg"
)

type ImageRepo struct {
	db *dbpg.DB
}

func New(db *dbpg.DB) *ImageRepo {
	return &ImageRepo{db: db}
}

func (r *ImageRepo) CreateImage(ctx context.Context, image domain.Image) error {
	const op = "repo.image.Create"

	query := `INSERT INTO images(id, original_path) VALUES ($1, $2)`

	if _, err := r.db.ExecContext(ctx, query, image.ID, image.Path); err != nil {
		return errutils.Wrap(op, err)
	}

	return nil
}

func (r *ImageRepo) GetImageByID(ctx context.Context, ID uuid.UUID) (domain.Image, error) {
	const op = "repo.image.Get"

	query := `
		SELECT id, original_path, uploaded_at
		FROM images
		WHERE id = $1
	`

	var image domain.Image
	if err := r.db.QueryRowContext(ctx, query, ID).Scan(
		&image.ID,
		&image.Path,
		&image.UploadedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Image{}, errutils.Wrap(op, repo.ErrImageNotFound)
		}
		return domain.Image{}, errutils.Wrap(op, err)
	}

	return image, nil
}

func (r *ImageRepo) GetPathByID(ctx context.Context, ID uuid.UUID) (string, error) {
	const op = "repo.image.GetPathByID"

	query := `SELECT original_path FROM images WHERE id = $1`

	var path string

	if err := r.db.QueryRowContext(ctx, query, ID).Scan(&path); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", repo.ErrImageNotFound
		}
		return "", err
	}

	return path, nil
}

func (r *ImageRepo) DeleteImageByID(ctx context.Context, ID uuid.UUID) error {
	const op = "repo.image.DeleteByID"

	query := `DELETE FROM images WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, ID)
	if err != nil {
		return errutils.Wrap(op, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errutils.Wrap(op, err)
	}

	if rows == 0 {
		return errutils.Wrap(op, repo.ErrImageNotFound)
	}

	return nil
}
