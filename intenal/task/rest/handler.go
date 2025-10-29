package rest

import (
	"context"
	"github.com/google/uuid"
	"github.com/ilam072/image-processor/intenal/response"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
	"net/http"
)

type Task interface {
	EnqueueTask(ctx context.Context, ID uuid.UUID, action string) (int, string, error)
}

type Handler struct {
	task Task
}

func NewTaskHandler(task Task) *Handler {
	return &Handler{task: task}
}

// GET /image/:id/tasks/resize
// GET /image/:id/tasks/thumbnail
// GET /image/:id/tasks/watermark
func (h *Handler) EnqueueTask(action string) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.Error("id must be UUID format")
			return
		}

		taskID, status, err := h.task.EnqueueTask(c.Request.Context(), id, action)
		if err != nil {
			zlog.Logger.Error().Err(err).Str("action", action).Str("image_id", id.String()).Msg("failed to enqueue task")
			response.Error("internal server error, try again later").WriteJSON(c, http.StatusInternalServerError)
			return
		}

		response.Raw(c, http.StatusOK, ginext.H{"task_id": taskID, "status": status})

	}
}
