package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/ilam072/image-processor/intenal/types/dto"
	"github.com/segmentio/kafka-go"
	"github.com/wb-go/wbf/zlog"
	"io"
)

type Consumer interface {
	Consume(ctx context.Context) (kafka.Message, error)
	Commit(ctx context.Context, msg kafka.Message) error
	Close() error
}

type Image interface {
	GetPathByID(ctx context.Context, ID uuid.UUID) (string, error)
}

type ImageProcessor interface {
	Resize(img io.ReadCloser) (io.ReadCloser, error)
	Watermark(img io.ReadCloser) (io.ReadCloser, error)
	Thumbnail(img io.ReadCloser) (io.ReadCloser, error)
}

type Task interface {
	GetTaskByID(ctx context.Context, id int) (dto.Task, error)
	UpdateTaskStatus(ctx context.Context, ID int, status string) error
}

type ImageStorage interface {
	Load(ctx context.Context, name string) (io.ReadCloser, error)
	SaveImage(ctx context.Context, name string, file io.Reader) error
}
type Handler struct {
	task      Task
	image     Image
	storage   ImageStorage
	processor ImageProcessor
	c         Consumer
}

func New(task Task, image Image, storage ImageStorage, processor ImageProcessor, c Consumer) *Handler {
	return &Handler{
		task:      task,
		image:     image,
		storage:   storage,
		processor: processor,
		c:         c,
	}
}

func (h *Handler) Start(ctx context.Context) error {
	const op = "kafka.handler.Start"

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := h.c.Consume(ctx)
			if err != nil {
				zlog.Logger.Warn().Err(err).Msg("failed to consume message")
				continue
			}

			fmt.Println("Получил message")

			var taskMsg dto.TaskMessage
			if err := json.Unmarshal(msg.Value, &taskMsg); err != nil {
				zlog.Logger.Warn().Int("task_id", taskMsg.ID).Msg("invalid task id format")
				_ = h.task.UpdateTaskStatus(ctx, taskMsg.ID, "failed")
				continue
			}

			fmt.Println("Анмаршаллил message")

			task, err := h.task.GetTaskByID(ctx, taskMsg.ID)
			if err != nil {
				zlog.Logger.Error().Err(err).Int("task_id", taskMsg.ID).Msg("failed to get task")
				_ = h.task.UpdateTaskStatus(ctx, taskMsg.ID, "failed")
				continue
			}

			if err := h.task.UpdateTaskStatus(ctx, task.ID, "processing"); err != nil {
				zlog.Logger.Error().Err(err).Int("task_id", task.ID).Msg("failed to update task status to processing")
				_ = h.task.UpdateTaskStatus(ctx, task.ID, "failed")
				continue
			}

			fmt.Println("Обновил таск статус")

			originalPath, err := h.image.GetPathByID(ctx, task.ImageID)
			if err != nil {
				zlog.Logger.Error().Err(err).Str("image_id", task.ImageID.String()).Msg("failed to get image path by id to processing")
				_ = h.task.UpdateTaskStatus(ctx, task.ID, "failed")
				continue
			}
			// image.Download(ctx, originalPath)
			img, err := h.storage.Load(ctx, originalPath)
			if err != nil {
				zlog.Logger.Error().Err(err).Str("original_path", originalPath).Msg("failed to get image path by id to processing")
				_ = h.task.UpdateTaskStatus(ctx, task.ID, "failed")
				continue
			}

			func() {
				defer img.Close()

				// Process
				processedImg, err := h.process(img, task.Type)
				if err != nil {
					zlog.Logger.Error().Err(err)
					_ = h.task.UpdateTaskStatus(ctx, task.ID, "failed")
					return
				}
				// Сохранить image по processed path
				if err = h.storage.SaveImage(ctx, task.ProcessedPath, processedImg); err != nil {
					zlog.Logger.Error().Err(err)
					_ = h.task.UpdateTaskStatus(ctx, task.ID, "failed")
					return
				}

				if err := h.task.UpdateTaskStatus(ctx, task.ID, "completed"); err != nil {
					zlog.Logger.Error().Err(err).Int("task_id", task.ID).Msg("failed to update task status to completed")
					_ = h.task.UpdateTaskStatus(ctx, task.ID, "failed")
					return
				}

				if err := h.c.Commit(ctx, msg); err != nil {
					zlog.Logger.Error().Err(err).Int("task_id", task.ID)
					_ = h.task.UpdateTaskStatus(ctx, task.ID, "failed")
					return
				}
			}()
		}

	}
}

func (h *Handler) process(img io.ReadCloser, action string) (io.ReadCloser, error) {
	switch action {
	case "resize":
		return h.processor.Resize(img)
	case "thumbnail":
		return h.processor.Thumbnail(img)
	case "watermark":
		return h.processor.Watermark(img)
	default:
		return nil, fmt.Errorf("unexpected action")
	}
}
