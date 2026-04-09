package http

import (
	"net/http"
	"strconv"

	"backend/internals/student/controller/dto"
	"backend/internals/student/usecase"
	"github.com/gin-gonic/gin"
)

type StudentHandler struct {
	examUseCase    usecase.IStudentExamUseCase
	resultsUseCase usecase.IStudentResultsUseCase
}

func NewStudentHandler(examUseCase usecase.IStudentExamUseCase, resultsUseCase usecase.IStudentResultsUseCase) *StudentHandler {
	return &StudentHandler{
		examUseCase:    examUseCase,
		resultsUseCase: resultsUseCase,
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
	studentID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	studentIDInt, _ := studentID.(int64)
	response, err := h.resultsUseCase.GetExamResults(c.Request.Context(), studentIDInt)
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

	response, err := h.resultsUseCase.GetClassRanking(c.Request.Context(), examID)
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
