package dto

import "github.com/google/uuid"

type Task struct {
	ID            int
	ImageID       uuid.UUID
	ProcessedPath string
	Type          string
	Status        string
}

type TaskMessage struct {
	ID int
}
