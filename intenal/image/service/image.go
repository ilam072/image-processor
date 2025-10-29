package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/ilam072/image-processor/intenal/image/repo"
	"github.com/ilam072/image-processor/intenal/types/domain"
	"github.com/ilam072/image-processor/pkg/errutils"
	"io"
)

type ImageRepo interface {
	CreateImage(ctx context.Context, image domain.Image) error
	GetPathByID(ctx context.Context, ID uuid.UUID) (string, error)
	DeleteImageByID(ctx context.Context, ID uuid.UUID) error
}

type ImageStorage interface {
	SaveImage(ctx context.Context, name string, file io.Reader) error
	DeleteImage(ctx context.Context, name string) error
}

type Image struct {
	repo    ImageRepo
	storage ImageStorage
}

func New(repo ImageRepo, storage ImageStorage) *Image {
	return &Image{repo: repo, storage: storage}
}

func (i *Image) UploadImage(ctx context.Context, file io.Reader, ext string) (string, error) {
	const op = "service.image.Upload"

	ID := uuid.New()
	name := fmt.Sprintf("original/%s%s", ID.String(), ext)
	if err := i.storage.SaveImage(ctx, name, file); err != nil {
		return "", errutils.Wrap(op, err)
	}

	image := domain.Image{
		ID:   ID,
		Path: name,
	}

	if err := i.repo.CreateImage(ctx, image); err != nil {
		_ = i.storage.DeleteImage(ctx, name)
		return "", errutils.Wrap(op, err)
	}

	return ID.String(), nil
}

func (i *Image) GetPathByID(ctx context.Context, ID uuid.UUID) (string, error) {
	const op = "service.image.GetPathByID"

	path, err := i.repo.GetPathByID(ctx, ID)
	if err != nil {
		if errors.Is(err, repo.ErrImageNotFound) {
			return "", errutils.Wrap(op, domain.ErrImageNotFound)
		}
		return "", errutils.Wrap(op, err)
	}

	return path, nil
}

func (i *Image) DeleteImage(ctx context.Context, ID uuid.UUID) error {
	const op = "service.image.Delete"

	path, err := i.repo.GetPathByID(ctx, ID)
	if err != nil {
		if errors.Is(err, repo.ErrImageNotFound) {
			return errutils.Wrap(op, domain.ErrImageNotFound)
		}
		return errutils.Wrap(op, err)
	}

	if err := i.storage.DeleteImage(ctx, path); err != nil {
		return errutils.Wrap(op, err)
	}

	if err := i.repo.DeleteImageByID(ctx, ID); err != nil {
		return errutils.Wrap(op, err)
	}

	return nil
}
