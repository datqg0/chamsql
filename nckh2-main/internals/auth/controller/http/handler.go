package http

import (
	"backend/internals/auth/controller/dto"
	"backend/internals/auth/usecase"
	"backend/pkgs/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	usecase usecase.IAuthUseCase
}

func NewAuthHandler(usecase usecase.IAuthUseCase) *AuthHandler {
	return &AuthHandler{usecase: usecase}
}

// Register godoc
// @Summary     Register a new user
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       request body dto.RegisterRequest true "Registration data"
// @Success     201 {object} response.Response
// @Failure     400 {object} response.Response
// @Router      /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.usecase.Register(c.Request.Context(), &req)
	if err != nil {
		if err == usecase.ErrEmailExists {
			response.BadRequest(c, "Email already registered")
			return
		}
		if err == usecase.ErrUsernameExists {
			response.BadRequest(c, "Username already taken")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}

	response.Created(c, result)
}

// Login godoc
// @Summary     Login user
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       request body dto.LoginRequest true "Login credentials"
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Failure     401 {object} response.Response
// @Router      /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.usecase.Login(c.Request.Context(), &req)
	if err != nil {
		if err == usecase.ErrInvalidCredentials {
			response.Unauthorized(c, "Invalid phone or password")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}

	// Set Refresh Token as HttpOnly Cookie (7 days)
	// SameSite=Lax for normal functioning, Strict for higher security
	c.SetCookie("refresh_token", result.RefreshToken, 7*24*3600, "/", "", false, true)

	// Omit refreshToken from response body if you want strict security
	// string refreshToken = result.RefreshToken
	// result.RefreshToken = "" (optional)

	response.Success(c, result)
}

// Logout godoc
// @Summary     Logout user
// @Tags        Auth
// @Produce     json
// @Success     200 {object} response.Response
// @Failure     401 {object} response.Response
// @Router      /auth/logout [post]
// @Security    ApiKeyAuth
func (h *AuthHandler) Logout(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		response.Unauthorized(c, "Missing authorization header")
		return
	}
	// Strip "Bearer " prefix if present
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	err := h.usecase.Logout(c.Request.Context(), tokenString)
	if err != nil {
		response.Unauthorized(c, "Invalid token")
		return
	}

	// Clear Refresh Token Cookie
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)

	response.Success(c, gin.H{"message": "Logged out successfully"})
}

// RefreshToken godoc
// @Summary     Refresh access token
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       request body dto.RefreshTokenRequest true "Refresh Token"
// @Success     200 {object} dto.AuthResponse
// @Failure     400 {object} response.Response
// @Failure     401 {object} response.Response
// @Router      /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// 1. Try get from Cookie
	refreshToken, err := c.Cookie("refresh_token")

	// 2. Fallback to Body
	var req dto.RefreshTokenRequest
	if err != nil || refreshToken == "" {
		if errBinding := c.ShouldBindJSON(&req); errBinding != nil {
			response.BadRequest(c, "Refresh token required in Cookie or Body")
			return
		}
		refreshToken = req.RefreshToken
	} else {
		// Populate req for usecase
		req.RefreshToken = refreshToken
	}

	result, err := h.usecase.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		// Clear cookie if invalid
		c.SetCookie("refresh_token", "", -1, "/", "", false, true)
		response.Unauthorized(c, "Invalid or expired refresh token")
		return
	}

	// Update Cookie with new refresh token
	c.SetCookie("refresh_token", result.RefreshToken, 7*24*3600, "/", "", false, true)

	response.Success(c, result)
}
