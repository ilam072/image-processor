package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/ilam072/image-processor/intenal/types/domain"
	"github.com/ilam072/image-processor/intenal/utils"
	"github.com/ilam072/image-processor/pkg/errutils"
	"strconv"
)

type ImageRepo interface {
	GetImageByID(ctx context.Context, ID uuid.UUID) (domain.Image, error)
}

type TaskRepo interface {
	CreateTask(ctx context.Context, task domain.Task) (int, error)
}

type Producer interface {
	Produce(ctx context.Context, taskID string) error
}

type Task struct {
	imageRepo ImageRepo
	taskRepo  TaskRepo
	p         Producer
}

func (t *Task) EnqueueResizeTask(ctx context.Context, ID uuid.UUID) error {
	const op = "service.image.Upload"

	image, err := t.imageRepo.GetImageByID(ctx, ID)
	if err != nil {
		if errors.Is(err, domain.ErrImageNotFound) {
			return errutils.Wrap(op, domain.ErrImageNotFound)
		}
		return errutils.Wrap(op, err)
	}

	processedPath := utils.BuildProcessedPath(image.Path, "resize")
	task := domain.Task{
		ImageID:       image.ID,
		ProcessedPath: processedPath,
		Type:          domain.Resize,
		Status:        domain.Queued,
	}

	taskID, err := t.taskRepo.CreateTask(ctx, task)
	if err != nil {
		return errutils.Wrap(op, err)
	}

	return t.p.Produce(ctx, strconv.Itoa(taskID))
}
