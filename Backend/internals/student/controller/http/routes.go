package http

import (
	"backend/db"
	"backend/internals/student/usecase"
	"backend/pkgs/redis"
	"github.com/gin-gonic/gin"
)

func Routes(rg *gin.RouterGroup, database *db.Database, cache redis.IRedis, authMiddleware gin.HandlerFunc) {
	examUC := usecase.NewStudentExamUseCase(database, cache)
	resultsUC := usecase.NewStudentResultsUseCase(database)
	practiceUC := usecase.NewPracticeUseCase(database)
	handler := NewStudentHandler(examUC, resultsUC, practiceUC)

	student := rg.Group("/student")
	student.Use(authMiddleware)
	{
		// EXAM ENDPOINTS
		student.POST("/exams/join", handler.JoinExam)
		student.POST("/exams/start", handler.StartExam)
		student.GET("/exams/:examID", handler.GetExam)
		student.GET("/exams/:examID/time-remaining", handler.GetTimeRemaining)
		student.GET("/exams/:examID/problems/:problemID", handler.GetProblem)
		student.POST("/exams/:examID/problems/:problemID/submit", handler.SubmitCode)
		student.POST("/exams/submit", handler.SubmitExam)

		student.GET("/results", handler.GetExamResults)
		student.GET("/results/:examID", handler.GetExamResultDetail)
		student.GET("/exams/:examID/ranking", handler.GetClassRanking)
		student.GET("/exams/:examID/analytics", handler.GetExamAnalytics)

		// PRACTICE ENDPOINTS
		student.GET("/practice/problems", handler.ListPublicProblems)
		student.GET("/practice/problems/:id", handler.GetPublicProblem)
		student.GET("/practice/problems/slug/:slug", handler.GetPublicProblemBySlug)
		student.POST("/practice/problems/:id/submit", handler.PracticeSubmitCode)
		student.GET("/practice/problems/:id/submissions", handler.ListPracticeSubmissions)
	}
}
