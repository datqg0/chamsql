package http

import (
	"net/http"
	"strconv"

	"backend/internals/student/controller/dto"
	"backend/internals/student/usecase"
	"github.com/gin-gonic/gin"
)

type StudentHandler struct {
	examUseCase     usecase.IStudentExamUseCase
	resultsUseCase  usecase.IStudentResultsUseCase
	practiceUseCase usecase.IPracticeUseCase
}

func NewStudentHandler(examUseCase usecase.IStudentExamUseCase, resultsUseCase usecase.IStudentResultsUseCase, practiceUseCase usecase.IPracticeUseCase) *StudentHandler {
	return &StudentHandler{
		examUseCase:     examUseCase,
		resultsUseCase:  resultsUseCase,
		practiceUseCase: practiceUseCase,
	}
}

func (h *StudentHandler) JoinExam(c *gin.Context) {
	var req dto.JoinExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	studentID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	studentIDInt, _ := studentID.(int64)
	response, err := h.examUseCase.JoinExam(c.Request.Context(), req.ExamID, studentIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *StudentHandler) StartExam(c *gin.Context) {
	var req dto.StartExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	studentID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	studentIDInt, _ := studentID.(int64)
	response, err := h.examUseCase.StartExam(c.Request.Context(), req.ExamID, studentIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *StudentHandler) GetExam(c *gin.Context) {
	examID, err := strconv.ParseInt(c.Param("examID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam id"})
		return
	}

	studentID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	studentIDInt, _ := studentID.(int64)
	response, err := h.examUseCase.GetExam(c.Request.Context(), examID, studentIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *StudentHandler) GetProblem(c *gin.Context) {
	examID, err := strconv.ParseInt(c.Param("examID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam id"})
		return
	}

	problemID, err := strconv.ParseInt(c.Param("problemID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem id"})
		return
	}

	studentID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	studentIDInt, _ := studentID.(int64)
	response, err := h.examUseCase.GetProblem(c.Request.Context(), examID, problemID, studentIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *StudentHandler) SubmitCode(c *gin.Context) {
	examID, err := strconv.ParseInt(c.Param("examID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam id"})
		return
	}

	problemID, err := strconv.ParseInt(c.Param("problemID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem id"})
		return
	}

	var req dto.SubmitCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	studentID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	studentIDInt, _ := studentID.(int64)
	response, err := h.examUseCase.SubmitCode(c.Request.Context(), examID, problemID, studentIDInt, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *StudentHandler) SubmitExam(c *gin.Context) {
	var req dto.SubmitExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	studentID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	studentIDInt, _ := studentID.(int64)
	response, err := h.examUseCase.SubmitExam(c.Request.Context(), req.ExamID, studentIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *StudentHandler) GetTimeRemaining(c *gin.Context) {
	examID, err := strconv.ParseInt(c.Param("examID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam id"})
		return
	}

	studentID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	studentIDInt, _ := studentID.(int64)
	response, err := h.examUseCase.GetTimeRemaining(c.Request.Context(), examID, studentIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *StudentHandler) GetExamResults(c *gin.Context) {
	var req dto.ListExamResultsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}

	studentID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	studentIDInt, _ := studentID.(int64)
	response, err := h.resultsUseCase.GetExamResults(c.Request.Context(), studentIDInt, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *StudentHandler) GetExamResultDetail(c *gin.Context) {
	examID, err := strconv.ParseInt(c.Param("examID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam id"})
		return
	}

	studentID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	studentIDInt, _ := studentID.(int64)
	response, err := h.resultsUseCase.GetExamResultDetail(c.Request.Context(), examID, studentIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *StudentHandler) GetClassRanking(c *gin.Context) {
	examID, err := strconv.ParseInt(c.Param("examID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam id"})
		return
	}

	var req dto.RankingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 50
	}

	response, err := h.resultsUseCase.GetClassRanking(c.Request.Context(), examID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *StudentHandler) GetExamAnalytics(c *gin.Context) {
	examID, err := strconv.ParseInt(c.Param("examID"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam id"})
		return
	}

	response, err := h.resultsUseCase.GetExamAnalytics(c.Request.Context(), examID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// =============================================
// PRACTICE ENDPOINTS
// =============================================

// ListPublicProblems - List all public problems for practice
func (h *StudentHandler) ListPublicProblems(c *gin.Context) {
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

	difficulty := c.Query("difficulty")
	topic := c.Query("topic")

	response, err := h.practiceUseCase.ListPublicProblems(c.Request.Context(), page, pageSize, difficulty, topic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetPublicProblem - Get full details of a public problem by ID
func (h *StudentHandler) GetPublicProblem(c *gin.Context) {
	problemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem id"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDInt, _ := userID.(int64)

	response, err := h.practiceUseCase.GetPublicProblem(c.Request.Context(), problemID, userIDInt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetPublicProblemBySlug - Get full details of a public problem by slug
func (h *StudentHandler) GetPublicProblemBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDInt, _ := userID.(int64)

	response, err := h.practiceUseCase.GetPublicProblemBySlug(c.Request.Context(), slug, userIDInt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// PracticeSubmitCode - Submit code for a practice problem
func (h *StudentHandler) PracticeSubmitCode(c *gin.Context) {
	problemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem id"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.PracticeSubmitCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDInt, _ := userID.(int64)

	response, err := h.practiceUseCase.PracticeSubmitCode(c.Request.Context(), problemID, userIDInt, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// ListPracticeSubmissions - List practice submissions for a problem
func (h *StudentHandler) ListPracticeSubmissions(c *gin.Context) {
	problemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem id"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

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

	userIDInt, _ := userID.(int64)

	response, err := h.practiceUseCase.ListPracticeSubmissions(c.Request.Context(), problemID, userIDInt, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetMySubmissions - Lịch sử nộp bài luyện tập tổng hợp
// GET /student/submissions?page=1&pageSize=20
func (h *StudentHandler) GetMySubmissions(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
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

	userIDInt, _ := userID.(int64)
	response, err := h.resultsUseCase.GetMySubmissions(c.Request.Context(), userIDInt, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
