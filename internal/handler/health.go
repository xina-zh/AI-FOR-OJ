package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ai-for-oj/internal/service"
)

type HealthHandler struct {
	service *service.HealthService
}

func NewHealthHandler(service *service.HealthService) *HealthHandler {
	return &HealthHandler{service: service}
}

func (h *HealthHandler) Get(c *gin.Context) {
	status := h.service.Status(c.Request.Context())
	statusCode := http.StatusOK
	if status.Status != "ok" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, status)
}
