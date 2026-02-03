package http

import (
	"strconv"

	"backend/internals/topic/controller/dto"
	"backend/internals/topic/usecase"
	"backend/pkgs/response"

	"github.com/gin-gonic/gin"
)

type TopicHandler struct {
	usecase usecase.ITopicUseCase
}

func NewTopicHandler(uc usecase.ITopicUseCase) *TopicHandler {
	return &TopicHandler{usecase: uc}
}

// List godoc
// @Summary     List all topics
// @Tags        Topics
// @Produce     json
// @Success     200 {object} dto.TopicListResponse
// @Router      /topics [get]
func (h *TopicHandler) List(c *gin.Context) {
	result, err := h.usecase.List(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// GetBySlug godoc
// @Summary     Get topic by slug
// @Tags        Topics
// @Produce     json
// @Param       slug path string true "Topic slug"
// @Success     200 {object} dto.TopicResponse
// @Router      /topics/{slug} [get]
func (h *TopicHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	result, err := h.usecase.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		if err == usecase.ErrTopicNotFound {
			response.NotFound(c, "Topic not found")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// Create godoc
// @Summary     Create a new topic
// @Tags        Topics
// @Accept      json
// @Produce     json
// @Param       request body dto.CreateTopicRequest true "Topic data"
// @Success     201 {object} dto.TopicResponse
// @Router      /topics [post]
func (h *TopicHandler) Create(c *gin.Context) {
	var req dto.CreateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.usecase.Create(c.Request.Context(), &req)
	if err != nil {
		if err == usecase.ErrSlugExists {
			response.BadRequest(c, "Topic slug already exists")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	response.Created(c, result)
}

// Update godoc
// @Summary     Update a topic
// @Tags        Topics
// @Accept      json
// @Produce     json
// @Param       id path int true "Topic ID"
// @Param       request body dto.UpdateTopicRequest true "Topic data"
// @Success     200 {object} dto.TopicResponse
// @Router      /topics/{id} [put]
func (h *TopicHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid topic ID")
		return
	}

	var req dto.UpdateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.usecase.Update(c.Request.Context(), int32(id), &req)
	if err != nil {
		if err == usecase.ErrTopicNotFound {
			response.NotFound(c, "Topic not found")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// Delete godoc
// @Summary     Delete a topic
// @Tags        Topics
// @Produce     json
// @Param       id path int true "Topic ID"
// @Success     200 {object} response.Response
// @Router      /topics/{id} [delete]
func (h *TopicHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid topic ID")
		return
	}

	err = h.usecase.Delete(c.Request.Context(), int32(id))
	if err != nil {
		if err == usecase.ErrTopicNotFound {
			response.NotFound(c, "Topic not found")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "Topic deleted successfully"})
}
