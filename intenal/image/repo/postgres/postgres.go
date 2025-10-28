package postgres

import (
	"context"
	"github.com/ilam072/image-processor/intenal/image/types/domain"
	"github.com/ilam072/image-processor/pkg/errutils"
	"github.com/wb-go/wbf/dbpg"
)

type Repo struct {
	db *dbpg.DB
}

func New(db *dbpg.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) CreateImage(ctx context.Context, image domain.Image) error {
	const op = "repo.image.Create"

	query := `INSERT INTO images(id, original_path) VALUES ($1, $2)`

	if _, err := r.db.ExecContext(ctx, query, image.ID, image.Path); err != nil {
		return errutils.Wrap(op, err)
	}

	return nil
}
