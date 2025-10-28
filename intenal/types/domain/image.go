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
