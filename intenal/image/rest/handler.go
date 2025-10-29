package rest

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/ilam072/image-processor/intenal/response"
	"github.com/ilam072/image-processor/intenal/types/domain"
	"github.com/ilam072/image-processor/intenal/utils"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
	"io"
	"mime"
	"net/http"
	"path/filepath"
)

type Image interface {
	UploadImage(ctx context.Context, file io.Reader, ext string) (string, error)
	GetPathByID(ctx context.Context, ID uuid.UUID) (string, error)
	DeleteImage(ctx context.Context, ID uuid.UUID) error
}

type Storage interface {
	Load(ctx context.Context, name string) (io.ReadCloser, error)
}

type Handler struct {
	image   Image
	storage Storage
}

func NewImageHandler(image Image, storage Storage) *Handler {
	return &Handler{image: image, storage: storage}
}

func (h *Handler) UploadImage(c *ginext.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error("failed to get file").WriteJSON(c, http.StatusBadRequest)
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		zlog.Logger.Error().
			Err(err).
			Int64("file_size", fileHeader.Size).
			Str("file_name", fileHeader.Filename).
			Str("content_type", fileHeader.Header.Get("Content-Type")).
			Msg("failed to open file")
		response.Error("internal server error, try again later").WriteJSON(c, http.StatusInternalServerError)
		return
	}
	defer file.Close()

	ext := filepath.Ext(fileHeader.Filename)
	ID, err := h.image.UploadImage(c.Request.Context(), file, ext)
	if err != nil {
		zlog.Logger.Error().
			Err(err).
			Int64("file_size", fileHeader.Size).
			Str("file_name", fileHeader.Filename).
			Msg("failed to upload file")
		response.Error("internal server error, try again later").WriteJSON(c, http.StatusInternalServerError)
		return
	}

	response.Raw(c, http.StatusOK, map[string]string{"image_id": ID})
}

func (h *Handler) GetImage(c *ginext.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error("id must be UUID format").WriteJSON(c, http.StatusBadRequest)
		return
	}

	action := c.Query("processed")
	if action != "" && action != "resize" && action != "thumbnail" && action != "watermark" {
		response.Error("unexpected action").WriteJSON(c, http.StatusBadRequest)
		return
	}

	path, err := h.image.GetPathByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrImageNotFound) {
			response.Error("image not found").WriteJSON(c, http.StatusNotFound)
			return
		}
		zlog.Logger.Error().Err(err).Str("image_id", id.String()).Msg("failed to get image path")
		response.Error("internal server error, try again later").WriteJSON(c, http.StatusInternalServerError)
		return
	}

	if action != "" {
		original := path
		path = utils.BuildProcessedPath(original, action)
	}

	img, err := h.storage.Load(c.Request.Context(), path)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("path", path).Msg("failed to load image from storage")
		response.Error("image processing in progress").WriteJSON(c, http.StatusAccepted)
		return
	}
	defer img.Close()

	filename := filepath.Base(path)
	contentType := mime.TypeByExtension(filepath.Ext(filename))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))
	c.DataFromReader(http.StatusOK, -1, contentType, img, nil)
}

func (h *Handler) DeleteImage(c *ginext.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error("id must be UUID format").WriteJSON(c, http.StatusBadRequest)
		return
	}

	if err := h.image.DeleteImage(c.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrImageNotFound) {
			response.Error("image not found").WriteJSON(c, http.StatusNotFound)
			return
		}
		zlog.Logger.Error().Err(err).Str("image_id", id.String()).Msg("failed to delete image")
		response.Error("internal server error, try again later").WriteJSON(c, http.StatusInternalServerError)
		return
	}

	response.Success("image deleted successfully").WriteJSON(c, http.StatusOK)
}
