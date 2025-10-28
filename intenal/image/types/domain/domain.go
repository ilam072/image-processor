package domain

import (
	"github.com/google/uuid"
	"time"
)

type Image struct {
	ID         uuid.UUID
	Path       string
	UploadedAt time.Time
}

type TaskType string

const (
	Resize    TaskType = "resize"
	Thumbnail TaskType = "thumbnail"
	Watermark TaskType = "watermark"
)

type TaskStatus string

const (
	Queued     TaskStatus = "queued"
	Processing TaskStatus = "processing"
	Completed  TaskStatus = "completed"
	Failed     TaskStatus = "failed"
)

type Task struct {
	ID            int
	ImageID       uuid.UUID
	ProcessedPath string
	Type          TaskType
	Status        TaskStatus
	CreatedAt     time.Time
}
