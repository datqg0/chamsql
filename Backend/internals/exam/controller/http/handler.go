package http

import (
	"strconv"

	"backend/internals/exam/controller/dto"
	"backend/internals/exam/usecase"
	"backend/pkgs/middlewares"
	"backend/pkgs/response"

	"github.com/gin-gonic/gin"
)

type ExamHandler struct {
	usecase usecase.IExamUseCase
}

func NewExamHandler(uc usecase.IExamUseCase) *ExamHandler {
	return &ExamHandler{usecase: uc}
}

// ============ LECTURER CRUD ============

// List godoc
// @Summary     List exams
// @Tags        Exams
// @Produce     json
// @Success     200 {object} dto.ExamListResponse
// @Router      /exams [get]
func (h *ExamHandler) List(c *gin.Context) {
	userID, _ := middlewares.GetUserID(c)
	role, _ := middlewares.GetUserRole(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	result, err := h.usecase.List(c.Request.Context(), userID, role, page, pageSize)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// GetByID godoc
// @Summary     Get exam by ID
// @Tags        Exams
// @Produce     json
// @Param       id path int true "Exam ID"
// @Success     200 {object} dto.ExamResponse
// @Router      /exams/{id} [get]
func (h *ExamHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid exam ID")
		return
	}

	result, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == usecase.ErrExamNotFound {
			response.NotFound(c, "Exam not found")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// Create godoc
// @Summary     Create exam
// @Tags        Exams
// @Accept      json
// @Produce     json
// @Param       request body dto.CreateExamRequest true "Exam data"
// @Success     201 {object} dto.ExamResponse
// @Router      /exams [post]
func (h *ExamHandler) Create(c *gin.Context) {
	userID, ok := middlewares.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	var req dto.CreateExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.usecase.Create(c.Request.Context(), userID, &req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Created(c, result)
}

// Update godoc
// @Summary     Update exam
// @Tags        Exams
// @Accept      json
// @Produce     json
// @Param       id path int true "Exam ID"
// @Param       request body dto.UpdateExamRequest true "Exam data"
// @Success     200 {object} dto.ExamResponse
// @Router      /exams/{id} [put]
func (h *ExamHandler) Update(c *gin.Context) {
	userID, ok := middlewares.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid exam ID")
		return
	}

	var req dto.UpdateExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.usecase.Update(c.Request.Context(), userID, id, &req)
	if err != nil {
		handleExamError(c, err)
		return
	}
	response.Success(c, result)
}

// Delete godoc
// @Summary     Delete exam
// @Tags        Exams
// @Produce     json
// @Param       id path int true "Exam ID"
// @Success     200 {object} response.Response
// @Router      /exams/{id} [delete]
func (h *ExamHandler) Delete(c *gin.Context) {
	userID, ok := middlewares.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid exam ID")
		return
	}

	err = h.usecase.Delete(c.Request.Context(), userID, id)
	if err != nil {
		handleExamError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Exam deleted successfully"})
}

// ============ PROBLEM MANAGEMENT ============

// AddProblem godoc
// @Summary     Add problem to exam
// @Tags        Exams
// @Accept      json
// @Produce     json
// @Param       id path int true "Exam ID"
// @Param       request body dto.AddProblemRequest true "Problem data"
// @Success     200 {object} response.Response
// @Router      /exams/{id}/problems [post]
func (h *ExamHandler) AddProblem(c *gin.Context) {
	userID, ok := middlewares.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	examID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid exam ID")
		return
	}

	var req dto.AddProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err = h.usecase.AddProblem(c.Request.Context(), userID, examID, &req)
	if err != nil {
		handleExamError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Problem added to exam"})
}

// RemoveProblem godoc
// @Summary     Remove problem from exam
// @Tags        Exams
// @Produce     json
// @Param       id path int true "Exam ID"
// @Param       problemId path int true "Problem ID"
// @Success     200 {object} response.Response
// @Router      /exams/{id}/problems/{problemId} [delete]
func (h *ExamHandler) RemoveProblem(c *gin.Context) {
	userID, ok := middlewares.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	examID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	problemID, _ := strconv.ParseInt(c.Param("problemId"), 10, 64)

	err := h.usecase.RemoveProblem(c.Request.Context(), userID, examID, problemID)
	if err != nil {
		handleExamError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Problem removed from exam"})
}

// ListProblems godoc
// @Summary     List exam problems
// @Tags        Exams
// @Produce     json
// @Param       id path int true "Exam ID"
// @Success     200 {array} dto.ExamProblemResponse
// @Router      /exams/{id}/problems [get]
func (h *ExamHandler) ListProblems(c *gin.Context) {
	examID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid exam ID")
		return
	}

	result, err := h.usecase.ListProblems(c.Request.Context(), examID)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// ============ PARTICIPANT MANAGEMENT ============

// AddParticipants godoc
// @Summary     Add participants to exam
// @Tags        Exams
// @Accept      json
// @Produce     json
// @Param       id path int true "Exam ID"
// @Param       request body dto.AddParticipantsRequest true "Participant IDs"
// @Success     200 {object} response.Response
// @Router      /exams/{id}/participants [post]
func (h *ExamHandler) AddParticipants(c *gin.Context) {
	userID, ok := middlewares.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	examID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req dto.AddParticipantsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err := h.usecase.AddParticipants(c.Request.Context(), userID, examID, &req)
	if err != nil {
		handleExamError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Participants added"})
}

// ListParticipants godoc
// @Summary     List exam participants
// @Tags        Exams
// @Produce     json
// @Param       id path int true "Exam ID"
// @Success     200 {array} dto.ParticipantResponse
// @Router      /exams/{id}/participants [get]
func (h *ExamHandler) ListParticipants(c *gin.Context) {
	examID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	result, err := h.usecase.ListParticipants(c.Request.Context(), examID)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// RemoveParticipant godoc
// @Summary     Remove participant from exam
// @Tags        Exams
// @Produce     json
// @Param       id path int true "Exam ID"
// @Param       userId path int true "User ID"
// @Success     200 {object} response.Response
// @Router      /exams/{id}/participants/{userId} [delete]
func (h *ExamHandler) RemoveParticipant(c *gin.Context) {
	userID, ok := middlewares.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	examID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	participantID, _ := strconv.ParseInt(c.Param("userId"), 10, 64)

	err := h.usecase.RemoveParticipant(c.Request.Context(), userID, examID, participantID)
	if err != nil {
		handleExamError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Participant removed"})
}

// ============ STUDENT ACTIONS ============

// StartExam godoc
// @Summary     Start exam (student)
// @Tags        Exams
// @Produce     json
// @Param       id path int true "Exam ID"
// @Success     200 {object} dto.StartExamResponse
// @Router      /exams/{id}/start [post]
func (h *ExamHandler) StartExam(c *gin.Context) {
	userID, ok := middlewares.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	examID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	result, err := h.usecase.StartExam(c.Request.Context(), userID, examID)
	if err != nil {
		handleExamError(c, err)
		return
	}
	response.Success(c, result)
}

// SubmitAnswer godoc
// @Summary     Submit answer for exam problem
// @Tags        Exams
// @Accept      json
// @Produce     json
// @Param       id path int true "Exam ID"
// @Param       request body dto.ExamSubmitRequest true "Answer data"
// @Success     200 {object} dto.ExamSubmitResponse
// @Router      /exams/{id}/submit [post]
func (h *ExamHandler) SubmitAnswer(c *gin.Context) {
	userID, ok := middlewares.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	examID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req dto.ExamSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.usecase.SubmitAnswer(c.Request.Context(), userID, examID, &req)
	if err != nil {
		handleExamError(c, err)
		return
	}
	response.Success(c, result)
}

// FinishExam godoc
// @Summary     Finish exam (student)
// @Tags        Exams
// @Produce     json
// @Param       id path int true "Exam ID"
// @Success     200 {object} dto.ExamResultResponse
// @Router      /exams/{id}/finish [post]
func (h *ExamHandler) FinishExam(c *gin.Context) {
	userID, ok := middlewares.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	examID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	result, err := h.usecase.FinishExam(c.Request.Context(), userID, examID)
	if err != nil {
		handleExamError(c, err)
		return
	}
	response.Success(c, result)
}

// GetMyExams godoc
// @Summary     Get my exams (student)
// @Tags        Exams
// @Produce     json
// @Success     200 {array} dto.ExamResponse
// @Router      /my-exams [get]
func (h *ExamHandler) GetMyExams(c *gin.Context) {
	userID, ok := middlewares.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	result, err := h.usecase.GetMyExams(c.Request.Context(), userID)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

func handleExamError(c *gin.Context, err error) {
	switch err {
	case usecase.ErrExamNotFound:
		response.NotFound(c, "Exam not found")
	case usecase.ErrNotParticipant:
		response.Forbidden(c, "You are not a participant of this exam")
	case usecase.ErrExamNotStarted:
		response.BadRequest(c, "Exam has not started yet")
	case usecase.ErrExamEnded:
		response.BadRequest(c, "Exam has ended")
	case usecase.ErrAlreadyStarted:
		response.BadRequest(c, "You have already started this exam")
	case usecase.ErrAlreadySubmitted:
		response.BadRequest(c, "You have already submitted this exam")
	case usecase.ErrMaxAttemptsReached:
		response.BadRequest(c, "Maximum attempts reached for this problem")
	case usecase.ErrTimeExpired:
		response.BadRequest(c, "Exam time has expired")
	case usecase.ErrProblemNotInExam:
		response.BadRequest(c, "Problem not found in this exam")
	case usecase.ErrUnauthorized:
		response.Forbidden(c, "You are not authorized to perform this action")
	default:
		response.InternalServerError(c, err.Error())
	}
}
