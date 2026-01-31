package http

import (
	"backend/configs"
	"backend/db"
	problemRepo "backend/internals/problem/repository"
	"backend/internals/submission/repository"
	"backend/internals/submission/usecase"
	"backend/pkgs/runner"

	"github.com/gin-gonic/gin"
)

func Routes(rg *gin.RouterGroup, database *db.Database, queryRunner runner.Runner, cfg *configs.Config, authMiddleware gin.HandlerFunc) {
	subRepo := repository.NewSubmissionRepository(database)
	probRepo := problemRepo.NewProblemRepository(database)
	uc := usecase.NewSubmissionUseCase(subRepo, probRepo, queryRunner, cfg)
	handler := NewSubmissionHandler(uc)

	// Problem submission routes
	problems := rg.Group("/problems")
	{
		// Run can be done without auth (for testing)
		problems.POST("/:problemId/run", handler.Run)

		// Submit requires auth
		protected := problems.Group("")
		protected.Use(authMiddleware)
		{
			protected.POST("/:problemId/submit", handler.Submit)
		}
	}

	// Submission history routes (all require auth)
	submissions := rg.Group("/submissions")
	submissions.Use(authMiddleware)
	{
		submissions.GET("", handler.List)
		submissions.GET("/:id", handler.GetByID)
	}
}
