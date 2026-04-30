package http

import (
	"backend/db"
	"backend/internals/lecturer/usecase"
	"backend/pkgs/redis"
	"github.com/gin-gonic/gin"
)

// Routes - Register all lecturer endpoints
// Requires authentication
func Routes(rg *gin.RouterGroup, database *db.Database, cache redis.IRedis, authMiddleware gin.HandlerFunc) {
	classUC := usecase.NewLecturerClassUseCase(database, cache)
	gradingUC := usecase.NewGradingUseCase(database)
	handler := NewLecturerHandler(classUC, gradingUC)

	lecturer := rg.Group("/lecturer")
	lecturer.Use(authMiddleware)
	{
		// CLASS MANAGEMENT ROUTES
		// POST /lecturer/classes - Create class (lecturer only)
		lecturer.POST("/classes", handler.CreateClass)

		// GET /lecturer/classes - List lecturer's classes
		lecturer.GET("/classes", handler.ListClasses)

		// GET /lecturer/classes/:id - Get class details
		lecturer.GET("/classes/:id", handler.GetClass)

		// PUT /lecturer/classes/:id - Update class
		lecturer.PUT("/lecturer/classes/:id", handler.UpdateClass)

		// DELETE /lecturer/classes/:id - Delete class
		lecturer.DELETE("/lecturer/classes/:id", handler.DeleteClass)

		// CLASS MEMBERS ROUTES
		// POST /lecturer/classes/:id/members - Add member to class
		lecturer.POST("/lecturer/classes/:id/members", handler.AddClassMember)

		// GET /lecturer/classes/:id/members - List class members
		lecturer.GET("/lecturer/classes/:id/members", handler.ListClassMembers)

		// DELETE /lecturer/classes/:id/members/:userId - Remove member from class
		lecturer.DELETE("/lecturer/classes/:id/members/:userId", handler.RemoveClassMember)

		// CLASS EXAMS ROUTES
		// POST /lecturer/classes/:id/exams - Assign exam to class
		lecturer.POST("/lecturer/classes/:id/exams", handler.AssignExamToClass)

		// GET /lecturer/classes/:id/exams - List class exams
		lecturer.GET("/lecturer/classes/:id/exams", handler.ListClassExams)

		// DELETE /lecturer/classes/:id/exams/:examId - Remove exam from class
		lecturer.DELETE("/lecturer/classes/:id/exams/:examId", handler.RemoveExamFromClass)

		// GRADING ROUTES
		// GET /lecturer/submissions/:submissionId - View submission for grading
		lecturer.GET("/submissions/:submissionId", handler.ViewSubmissionForGrading)

		// POST /lecturer/submissions/:submissionId/grade - Grade a submission
		lecturer.POST("/submissions/:submissionId/grade", handler.GradeSubmission)

		// GET /lecturer/exams/:examId/ungraded - List ungraded submissions for exam
		lecturer.GET("/exams/:examId/ungraded", handler.ListUngradedSubmissions)

		// GET /lecturer/exams/:examId/grading-stats - Get grading statistics for exam
		lecturer.GET("/exams/:examId/grading-stats", handler.GetExamGradingStats)

		// GET /lecturer/exams/:examId/results - Xem kết quả kỳ thi (điểm, rank sinh viên)
		lecturer.GET("/exams/:examId/results", handler.GetExamResults)

		// POST /lecturer/submissions/bulk-grade - Grade multiple submissions
		lecturer.POST("/submissions/bulk-grade", handler.BulkGradeSubmissions)
	}
}
