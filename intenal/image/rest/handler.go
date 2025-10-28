package rest

import (
	"context"
	"github.com/ilam072/image-processor/intenal/response"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
	"io"
	"net/http"
)

type Image interface {
	UploadImage(ctx context.Context, file io.Reader) (string, error)
}

type Handler struct {
	image Image
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

	ID, err := h.image.UploadImage(c.Request.Context(), file)
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
