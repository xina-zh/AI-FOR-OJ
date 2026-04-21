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
	aiHandler *handler.AIHandler,
	metaHandler *handler.MetaHandler,
	experimentHandler *handler.ExperimentHandler,
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
		router.DELETE("/api/v1/problems/:id", problemHandler.Delete)
		router.POST("/api/v1/problems/:id/testcases", problemHandler.CreateTestCase)
		router.GET("/api/v1/problems/:id/testcases", problemHandler.ListTestCases)
	}
	if submissionHandler != nil {
		router.GET("/api/v1/submissions", submissionHandler.List)
		router.GET("/api/v1/submissions/stats/problems", submissionHandler.AggregateByProblem)
		router.GET("/api/v1/submissions/:id", submissionHandler.Get)
		router.POST("/api/v1/submissions/judge", submissionHandler.Judge)
	}
	if aiHandler != nil {
		router.POST("/api/v1/ai/solve", aiHandler.Solve)
		router.GET("/api/v1/ai/solve-runs/:id", aiHandler.GetRun)
	}
	if metaHandler != nil {
		router.GET("/api/v1/meta/experiment-options", metaHandler.ExperimentOptions)
	}
	if experimentHandler != nil {
		router.POST("/api/v1/experiments/compare", experimentHandler.Compare)
		router.GET("/api/v1/experiments/compare", experimentHandler.ListCompare)
		router.GET("/api/v1/experiments/compare/:id", experimentHandler.GetCompare)
		router.POST("/api/v1/experiments/repeat", experimentHandler.Repeat)
		router.GET("/api/v1/experiments/repeat/:id", experimentHandler.GetRepeat)
		router.POST("/api/v1/experiments/run", experimentHandler.Run)
		router.GET("/api/v1/experiments/:id", experimentHandler.Get)
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
