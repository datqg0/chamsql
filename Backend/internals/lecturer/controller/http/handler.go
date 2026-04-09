package http

import (
	"errors"
	"net/http"
	"strconv"

	"backend/internals/lecturer/controller/dto"
	"backend/internals/lecturer/usecase"
	"github.com/gin-gonic/gin"
)

// LecturerHandler - HTTP handler for lecturer operations
type LecturerHandler struct {
	classUseCase   usecase.ILecturerClassUseCase
	gradingUseCase usecase.IGradingUseCase
}

// NewLecturerHandler - Create new lecturer handler
func NewLecturerHandler(classUseCase usecase.ILecturerClassUseCase, gradingUseCase usecase.IGradingUseCase) *LecturerHandler {
	return &LecturerHandler{
		classUseCase:   classUseCase,
		gradingUseCase: gradingUseCase,
	}
}

// =============================================
// CLASS MANAGEMENT HANDLERS
// =============================================

// CreateClass - Create new class
// @Summary Create new class
// @Description Create new class for lecturer
// @Tags Classes
// @Accept json
// @Produce json
// @Param request body dto.CreateClassRequest true "Create class request"
// @Success 201 {object} dto.ClassResponse
// @Failure 400 {string} string "Invalid request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /lecturer/classes [post]
func (h *LecturerHandler) CreateClass(c *gin.Context) {
	var req dto.CreateClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get lecturer ID from context (set by auth middleware)
	lecturerID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	lectureIDInt, _ := lecturerID.(int64)
	response, err := h.classUseCase.CreateClass(c.Request.Context(), lectureIDInt, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetClass - Get class by ID
// @Summary Get class details
// @Description Get class details by ID
// @Tags Classes
// @Produce json
// @Param id path int64 true "Class ID"
// @Success 200 {object} dto.ClassResponse
// @Failure 404 {string} string "Class not found"
// @Failure 401 {string} string "Unauthorized"
// @Router /lecturer/classes/{id} [get]
func (h *LecturerHandler) GetClass(c *gin.Context) {
	classID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid class id"})
		return
	}

	response, err := h.classUseCase.GetClass(c.Request.Context(), classID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListClasses - List lecturer's classes
// @Summary List classes
// @Description List all classes created by lecturer
// @Tags Classes
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Records per page" default(10)
// @Success 200 {object} dto.ListClassesResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /lecturer/classes [get]
func (h *LecturerHandler) ListClasses(c *gin.Context) {
	// Get lecturer ID from context
	lecturerID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	lectureIDInt, _ := lecturerID.(int64)

	page := 1
	pageSize := 10

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	response, err := h.classUseCase.ListClasses(c.Request.Context(), lectureIDInt, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateClass - Update class details
// @Summary Update class
// @Description Update class details
// @Tags Classes
// @Accept json
// @Produce json
// @Param id path int64 true "Class ID"
// @Param request body dto.UpdateClassRequest true "Update class request"
// @Success 200 {object} dto.ClassResponse
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Class not found"
// @Failure 401 {string} string "Unauthorized"
// @Router /lecturer/classes/{id} [put]
func (h *LecturerHandler) UpdateClass(c *gin.Context) {
	classID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid class id"})
		return
	}

	var req dto.UpdateClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.classUseCase.UpdateClass(c.Request.Context(), classID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteClass - Delete class
// @Summary Delete class
// @Description Delete class by ID
// @Tags Classes
// @Param id path int64 true "Class ID"
// @Success 204
// @Failure 404 {string} string "Class not found"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Router /lecturer/classes/{id} [delete]
func (h *LecturerHandler) DeleteClass(c *gin.Context) {
	classID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid class id"})
		return
	}

	// Get lecturer ID from context
	lecturerID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	lectureIDInt, _ := lecturerID.(int64)

	err = h.classUseCase.DeleteClass(c.Request.Context(), classID, lectureIDInt)
	if err != nil {
		if errors.Is(err, errors.New("only class creator can delete")) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// =============================================
// CLASS MEMBER HANDLERS
// =============================================

// AddClassMember - Add student to class
// @Summary Add class member
// @Description Add student to class
// @Tags Class Members
// @Accept json
// @Produce json
// @Param id path int64 true "Class ID"
// @Param request body dto.AddClassMemberRequest true "Add member request"
// @Success 201 {object} dto.ClassMemberResponse
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Not found"
// @Failure 409 {string} string "User already in class"
// @Router /lecturer/classes/{id}/members [post]
func (h *LecturerHandler) AddClassMember(c *gin.Context) {
	classID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid class id"})
		return
	}

	var req dto.AddClassMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.classUseCase.AddClassMember(c.Request.Context(), classID, req.UserID, req.Role)
	if err != nil {
		if errors.Is(err, errors.New("user already in class")) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// ListClassMembers - List class members
// @Summary List class members
// @Description List all members in class
// @Tags Class Members
// @Produce json
// @Param id path int64 true "Class ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Records per page" default(20)
// @Success 200 {object} dto.ListClassMembersResponse
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Class not found"
// @Router /lecturer/classes/{id}/members [get]
func (h *LecturerHandler) ListClassMembers(c *gin.Context) {
	classID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid class id"})
		return
	}

	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	response, err := h.classUseCase.ListClassMembers(c.Request.Context(), classID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RemoveClassMember - Remove student from class
// @Summary Remove class member
// @Description Remove student from class
// @Tags Class Members
// @Param id path int64 true "Class ID"
// @Param userId path int64 true "User ID"
// @Success 204
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Not found"
// @Router /lecturer/classes/{id}/members/{userId} [delete]
func (h *LecturerHandler) RemoveClassMember(c *gin.Context) {
	classID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid class id"})
		return
	}

	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	err = h.classUseCase.RemoveClassMember(c.Request.Context(), classID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// =============================================
// CLASS EXAM HANDLERS
// =============================================

// AssignExamToClass - Assign exam to class
// @Summary Assign exam to class
// @Description Assign exam to class for students
// @Tags Class Exams
// @Accept json
// @Produce json
// @Param id path int64 true "Class ID"
// @Param request body dto.AssignExamToClassRequest true "Assign exam request"
// @Success 201
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Not found"
// @Failure 409 {string} string "Exam already assigned"
// @Router /lecturer/classes/{id}/exams [post]
func (h *LecturerHandler) AssignExamToClass(c *gin.Context) {
	classID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid class id"})
		return
	}

	var req dto.AssignExamToClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.classUseCase.AssignExamToClass(c.Request.Context(), classID, req.ExamID)
	if err != nil {
		if errors.Is(err, errors.New("exam already assigned to class")) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

// ListClassExams - List exams assigned to class
// @Summary List class exams
// @Description List all exams assigned to class
// @Tags Class Exams
// @Produce json
// @Param id path int64 true "Class ID"
// @Success 200 {object} dto.ListClassExamsResponse
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Class not found"
// @Router /lecturer/classes/{id}/exams [get]
func (h *LecturerHandler) ListClassExams(c *gin.Context) {
	classID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid class id"})
		return
	}

	response, err := h.classUseCase.ListClassExams(c.Request.Context(), classID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RemoveExamFromClass - Remove exam from class
// @Summary Remove class exam
// @Description Remove exam from class
// @Tags Class Exams
// @Param id path int64 true "Class ID"
// @Param examId path int64 true "Exam ID"
// @Success 204
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Not found"
// @Router /lecturer/classes/{id}/exams/{examId} [delete]
func (h *LecturerHandler) RemoveExamFromClass(c *gin.Context) {
	classID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid class id"})
		return
	}

	examID, err := strconv.ParseInt(c.Param("examId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam id"})
		return
	}

	err = h.classUseCase.RemoveExamFromClass(c.Request.Context(), classID, examID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// =============================================
// GRADING / SCORING HANDLERS
// =============================================

// GradeSubmission - Grade a single exam submission
// @Summary Grade exam submission
// @Description Grade a student's exam submission with a score and optional feedback
// @Tags Grading
// @Accept json
// @Produce json
// @Param submissionId path int64 true "Submission ID"
// @Param request body dto.GradeSubmissionRequest true "Grading request"
// @Success 200 {object} dto.SubmissionGradingResponse
// @Failure 400 {string} string "Invalid request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Submission not found"
// @Router /lecturer/submissions/{submissionId}/grade [post]
func (h *LecturerHandler) GradeSubmission(c *gin.Context) {
	submissionID, err := strconv.ParseInt(c.Param("submissionId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid submission id"})
		return
	}

	var req dto.GradeSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lecturerID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	lecturerIDInt, _ := lecturerID.(int64)
	req.SubmissionID = submissionID

	response, err := h.gradingUseCase.GradeSubmission(c.Request.Context(), submissionID, lecturerIDInt, &req)
	if err != nil {
		if errors.Is(err, errors.New("not found")) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

// ViewSubmissionForGrading - Get submission details for grading
// @Summary View submission for grading
// @Description Retrieve full submission details including code, outputs, and answers
// @Tags Grading
// @Produce json
// @Param submissionId path int64 true "Submission ID"
// @Success 200 {object} dto.ViewSubmissionResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Submission not found"
// @Router /lecturer/submissions/{submissionId} [get]
func (h *LecturerHandler) ViewSubmissionForGrading(c *gin.Context) {
	submissionID, err := strconv.ParseInt(c.Param("submissionId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid submission id"})
		return
	}

	lecturerID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	lecturerIDInt, _ := lecturerID.(int64)

	response, err := h.gradingUseCase.ViewSubmissionForGrading(c.Request.Context(), submissionID, lecturerIDInt)
	if err != nil {
		if errors.Is(err, errors.New("not found")) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListUngradedSubmissions - List ungraded submissions for an exam
// @Summary List ungraded submissions
// @Description Get all submissions that need grading for a specific exam
// @Tags Grading
// @Produce json
// @Param examId path int64 true "Exam ID"
// @Success 200 {object} dto.ListUngradedSubmissionsResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Exam not found"
// @Router /lecturer/exams/{examId}/ungraded [get]
func (h *LecturerHandler) ListUngradedSubmissions(c *gin.Context) {
	examID, err := strconv.ParseInt(c.Param("examId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam id"})
		return
	}

	lecturerID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	lecturerIDInt, _ := lecturerID.(int64)

	response, err := h.gradingUseCase.ListUngradedSubmissions(c.Request.Context(), examID, lecturerIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetExamGradingStats - Get grading statistics for an exam
// @Summary Get exam grading stats
// @Description Retrieve statistics on grading progress for an exam
// @Tags Grading
// @Produce json
// @Param examId path int64 true "Exam ID"
// @Success 200 {object} dto.ExamGradingStatsResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Exam not found"
// @Router /lecturer/exams/{examId}/grading-stats [get]
func (h *LecturerHandler) GetExamGradingStats(c *gin.Context) {
	examID, err := strconv.ParseInt(c.Param("examId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam id"})
		return
	}

	lecturerID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	lecturerIDInt, _ := lecturerID.(int64)

	response, err := h.gradingUseCase.GetExamGradingStats(c.Request.Context(), examID, lecturerIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// BulkGradeSubmissions - Grade multiple submissions at once
// @Summary Bulk grade submissions
// @Description Grade multiple submissions in a single request
// @Tags Grading
// @Accept json
// @Produce json
// @Param request body dto.BulkGradeRequest true "Bulk grading request"
// @Success 200 {object} dto.BulkGradeResponse
// @Failure 400 {string} string "Invalid request"
// @Failure 401 {string} string "Unauthorized"
// @Router /lecturer/submissions/bulk-grade [post]
func (h *LecturerHandler) BulkGradeSubmissions(c *gin.Context) {
	var req dto.BulkGradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lecturerID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	lecturerIDInt, _ := lecturerID.(int64)

	response, err := h.gradingUseCase.BulkGradeSubmissions(c.Request.Context(), lecturerIDInt, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
