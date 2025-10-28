package minio

import (
	"context"
	"github.com/ilam072/image-processor/pkg/errutils"
	"github.com/minio/minio-go/v7"
	"io"
)

type Storage struct {
	mc     *minio.Client
	bucket string
}

func New(client *minio.Client) *Storage {
	return &Storage{mc: client}
}

func (s *Storage) SaveImage(ctx context.Context, name string, file io.Reader) error {
	const op = "filestorage.image.Save"

	_, err := s.mc.PutObject(ctx, s.bucket, name, file, -1, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return errutils.Wrap(op, err)
	}

	return nil
}

func (s *Storage) DeleteImage(ctx context.Context, name string) error {
	const op = "filestorage.image.Delete"

	if err := s.mc.RemoveObject(ctx, s.bucket, name, minio.RemoveObjectOptions{}); err != nil {
		return errutils.Wrap(op, err)
	}

	return nil
}
