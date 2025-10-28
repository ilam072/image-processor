package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/ilam072/image-processor/intenal/types/domain"
	"github.com/ilam072/image-processor/pkg/errutils"
	"io"
)

type ImageRepo interface {
	CreateImage(ctx context.Context, image domain.Image) error
}

type ImageStorage interface {
	SaveImage(ctx context.Context, name string, file io.Reader) error
	DeleteImage(ctx context.Context, name string) error
}

type Image struct {
	imageRepo ImageRepo
	storage   ImageStorage
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

	if err := i.imageRepo.CreateImage(ctx, image); err != nil {
		_ = i.storage.DeleteImage(ctx, name)
		return "", errutils.Wrap(op, err)
	}

	return ID.String(), nil
}
