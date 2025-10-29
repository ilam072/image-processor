package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/ilam072/image-processor/intenal/task/repo"
	"github.com/ilam072/image-processor/intenal/types/domain"
	"github.com/ilam072/image-processor/intenal/types/dto"
	"github.com/ilam072/image-processor/intenal/utils"
	"github.com/ilam072/image-processor/pkg/errutils"
)

type ImageRepo interface {
	GetImageByID(ctx context.Context, ID uuid.UUID) (domain.Image, error)
}

type TaskRepo interface {
	CreateTask(ctx context.Context, task domain.Task) (int, error)
	GetTaskByID(ctx context.Context, id int) (domain.Task, error)
	UpdateTaskStatus(ctx context.Context, ID int, status domain.TaskStatus) error
}

type Producer interface {
	Produce(ctx context.Context, taskID int) error
}

type Task struct {
	imageRepo ImageRepo
	taskRepo  TaskRepo
	p         Producer
}

func New(imageRepo ImageRepo, taskRepo TaskRepo, p Producer) *Task {
	return &Task{imageRepo: imageRepo, taskRepo: taskRepo, p: p}
}

func (t *Task) EnqueueTask(ctx context.Context, ID uuid.UUID, action string) (int, string, error) {
	const op = "service.task.EnqueueTask"

	image, err := t.imageRepo.GetImageByID(ctx, ID)
	if err != nil {
		if errors.Is(err, domain.ErrImageNotFound) {
			return 0, "", errutils.Wrap(op, err)
		}
		return 0, "", errutils.Wrap(op, err)
	}

	processedPath := utils.BuildProcessedPath(image.Path, action)
	task := domain.Task{
		ImageID:       image.ID,
		ProcessedPath: processedPath,
		Type:          domain.TaskType(action),
		Status:        domain.Queued,
	}

	taskID, err := t.taskRepo.CreateTask(ctx, task)
	if err != nil {
		return 0, "", errutils.Wrap(op, err)
	}

	if err := t.p.Produce(ctx, taskID); err != nil {
		return 0, "", errutils.Wrap(op, err)
	}

	return taskID, string(task.Status), nil
}

func (t *Task) GetTaskByID(ctx context.Context, id int) (dto.Task, error) {
	const op = "service.task.GetTaskByID"

	task, err := t.taskRepo.GetTaskByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrTaskNotFound) {
			return dto.Task{}, errutils.Wrap(op, domain.ErrTaskNotFound)
		}
		return dto.Task{}, errutils.Wrap(op, err)
	}

	return domainToDto(task), nil
}

func (t *Task) UpdateTaskStatus(ctx context.Context, ID int, status string) error {
	const op = "service.task.UpdateTaskStatus"

	domainStatus := domain.TaskStatus(status)
	if err := t.taskRepo.UpdateTaskStatus(ctx, ID, domainStatus); err != nil {
		return errutils.Wrap(op, err)
	}

	return nil
}

func domainToDto(task domain.Task) dto.Task {
	return dto.Task{
		ID:            task.ID,
		ImageID:       task.ImageID,
		ProcessedPath: task.ProcessedPath,
		Type:          string(task.Type),
		Status:        string(task.Status),
	}
}
