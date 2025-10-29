package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/ilam072/image-processor/intenal/task/repo"
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

func (r *TaskRepo) GetTaskByID(ctx context.Context, id int) (domain.Task, error) {
	const op = "repo.task.GetTaskByID"

	query := `
		SELECT id, image_id, processed_path, type, status, created_at
		FROM tasks
		WHERE id = $1
	`

	var task domain.Task
	row := r.db.QueryRowContext(ctx, query, id)
	if err := row.Scan(
		&task.ID,
		&task.ImageID,
		&task.ProcessedPath,
		&task.Type,
		&task.Status,
		&task.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Task{}, errutils.Wrap(op, repo.ErrTaskNotFound)
		}
		return domain.Task{}, errutils.Wrap(op, err)
	}

	return task, nil
}

func (r *TaskRepo) UpdateTaskStatus(ctx context.Context, ID int, status domain.TaskStatus) error {
	const op = "repo.task.UpdateTaskStatus"

	query := `UPDATE tasks SET status = $1 WHERE id = $2`

	if _, err := r.db.ExecContext(ctx, query, status, ID); err != nil {
		return errutils.Wrap(op, err)
	}

	return nil
}
