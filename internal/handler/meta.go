package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ai-for-oj/internal/agent"
	"ai-for-oj/internal/handler/dto"
	"ai-for-oj/internal/prompt"
)

type MetaHandler struct {
	defaultModel string
}

func NewMetaHandler(defaultModel string) *MetaHandler {
	return &MetaHandler{defaultModel: defaultModel}
}

func (h *MetaHandler) ExperimentOptions(c *gin.Context) {
	c.JSON(http.StatusOK, dto.ExperimentOptionsResponse{
		DefaultModel: h.defaultModel,
		Prompts:      toExperimentOptions(prompt.ListSolvePrompts()),
		Agents:       toExperimentOptions(agent.ListSolveAgents()),
	})
}

func toExperimentOptions(names []string) []dto.ExperimentOptionResponse {
	options := make([]dto.ExperimentOptionResponse, 0, len(names))
	for _, name := range names {
		options = append(options, dto.ExperimentOptionResponse{
			Name:  name,
			Label: name,
		})
	}
	return options
}
