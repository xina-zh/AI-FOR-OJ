package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ai-for-oj/internal/handler/dto"
)

func parseUintParam(c *gin.Context, key string) (uint, bool) {
	raw := c.Param(key)
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return 0, false
	}
	return uint(value), true
}

func parsePositiveIntQuery(c *gin.Context, key string, fallback int) (int, bool) {
	raw := c.Query(key)
	if raw == "" {
		return fallback, true
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid " + key})
		return 0, false
	}
	return value, true
}

func parseOptionalUintQuery(c *gin.Context, key string) (*uint, bool) {
	raw := c.Query(key)
	if raw == "" {
		return nil, true
	}

	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid " + key})
		return nil, false
	}

	uintValue := uint(value)
	return &uintValue, true
}
