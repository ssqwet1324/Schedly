package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"notif_service/internal/entity"
	service "notif_service/internal/service"
	"notif_service/internal/worker"
)

type Handler struct {
	service *service.Service
	worker  *worker.Worker
}

func NewHandler(service *service.Service, worker *worker.Worker) *Handler {
	return &Handler{service: service, worker: worker}
}

// CreateTaskHandler - handler для создания уведомления
func (h *Handler) CreateTaskHandler(ctx *gin.Context) {
	var task entity.Task
	if err := ctx.ShouldBindJSON(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdTask, err := h.service.CreateTask(ctx, task)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	h.worker.RedefinitionWorker()
	ctx.JSON(http.StatusCreated, createdTask)
}

// DeleteTaskHandler - handler для удаления уведомления
func (h *Handler) DeleteTaskHandler(ctx *gin.Context) {
	var task entity.Task
	if err := ctx.ShouldBindJSON(&task); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.DeleteTask(ctx.Request.Context(), task.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}
	h.worker.RedefinitionWorker()
	ctx.JSON(http.StatusOK, gin.H{"message": "Delete task successfully"})

}
