package runtime

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"ai-for-oj/internal/config"
	"ai-for-oj/internal/handler"
)

func NewRouter(
	cfg config.Config,
	logger *slog.Logger,
	healthHandler *handler.HealthHandler,
	problemHandler *handler.ProblemHandler,
	submissionHandler *handler.SubmissionHandler,
) *gin.Engine {
	if cfg.App.Env != "local" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogMiddleware(logger))

	router.GET("/health", healthHandler.Get)
	if problemHandler != nil {
		router.POST("/api/v1/problems", problemHandler.Create)
		router.GET("/api/v1/problems", problemHandler.List)
		router.GET("/api/v1/problems/:id", problemHandler.Get)
		router.POST("/api/v1/problems/:id/testcases", problemHandler.CreateTestCase)
		router.GET("/api/v1/problems/:id/testcases", problemHandler.ListTestCases)
	}
	if submissionHandler != nil {
		router.GET("/api/v1/submissions", submissionHandler.List)
		router.GET("/api/v1/submissions/stats/problems", submissionHandler.AggregateByProblem)
		router.GET("/api/v1/submissions/:id", submissionHandler.Get)
		router.POST("/api/v1/submissions/judge", submissionHandler.Judge)
	}

	return router
}

func requestLogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		logger.Info("http request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"client_ip", c.ClientIP(),
		)
	}
}
