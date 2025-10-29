package domain

import "errors"

var (
	ErrImageNotFound = errors.New("image not found")
	ErrTaskNotFound  = errors.New("task not found")
)
