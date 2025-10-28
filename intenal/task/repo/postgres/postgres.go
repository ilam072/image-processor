package postgres

import (
	"context"
	"github.com/ilam072/image-processor/intenal/types/domain"
	"github.com/ilam072/image-processor/pkg/errutils"
	"github.com/wb-go/wbf/dbpg"
)

type TaskRepo struct {
	db *dbpg.DB
}

func New(db *dbpg.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

func (r *TaskRepo) CreateTask(ctx context.Context, task domain.Task) (int, error) {
	const op = "repo.task.Create"

	query := `
		INSERT INTO tasks (image_id, processed_path, type, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	row := r.db.QueryRowContext(ctx, query,
		task.ImageID,
		task.ProcessedPath,
		task.Type,
		task.Status,
	)

	var ID int
	if err := row.Scan(&ID); err != nil {
		return 0, errutils.Wrap(op, err)
	}

	return ID, nil
}
