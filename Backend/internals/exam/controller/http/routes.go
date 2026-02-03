package http

import (
	"backend/configs"
	"backend/db"
	"backend/internals/exam/repository"
	"backend/internals/exam/usecase"
	problemRepo "backend/internals/problem/repository"
	"backend/pkgs/middlewares"
	"backend/pkgs/runner"

	"github.com/gin-gonic/gin"
)

func Routes(rg *gin.RouterGroup, database *db.Database, queryRunner runner.Runner, cfg *configs.Config, authMiddleware gin.HandlerFunc) {
	examRepoImpl := repository.NewExamRepository(database)
	probRepoImpl := problemRepo.NewProblemRepository(database)
	uc := usecase.NewExamUseCase(examRepoImpl, probRepoImpl, queryRunner, cfg)
	handler := NewExamHandler(uc)

	exams := rg.Group("/exams")
	exams.Use(authMiddleware)
	{
		// List (all authenticated users, filtered by role)
		exams.GET("", handler.List)
		exams.GET("/:id", handler.GetByID)

		// Lecturer/Admin only
		lecturerRoutes := exams.Group("")
		lecturerRoutes.Use(middlewares.RoleMiddleware("lecturer", "admin"))
		{
			lecturerRoutes.POST("", handler.Create)
			lecturerRoutes.PUT("/:id", handler.Update)
			lecturerRoutes.DELETE("/:id", handler.Delete)

			// Problem management
			lecturerRoutes.GET("/:id/problems", handler.ListProblems)
			lecturerRoutes.POST("/:id/problems", handler.AddProblem)
			lecturerRoutes.DELETE("/:id/problems/:problemId", handler.RemoveProblem)

			// Participant management
			lecturerRoutes.GET("/:id/participants", handler.ListParticipants)
			lecturerRoutes.POST("/:id/participants", handler.AddParticipants)
			lecturerRoutes.DELETE("/:id/participants/:userId", handler.RemoveParticipant)
		}

		// Student actions
		exams.POST("/:id/start", handler.StartExam)
		exams.POST("/:id/submit", handler.SubmitAnswer)
		exams.POST("/:id/finish", handler.FinishExam)
	}

	// Student's exam list
	rg.GET("/my-exams", authMiddleware, handler.GetMyExams)
}
