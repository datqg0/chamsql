package http

import (
	"context"
	"net/http"

	"backend/configs"
	"backend/pkgs/middlewares"
	"backend/pkgs/response"
	"backend/pkgs/runner"

	"github.com/gin-gonic/gin"
)

// SandboxHandler handles sandbox testing and monitoring
type SandboxHandler struct {
	cfg    *configs.Config
	runner runner.Runner
}

// NewSandboxHandler creates a new sandbox handler
func NewSandboxHandler(cfg *configs.Config, r runner.Runner) *SandboxHandler {
	return &SandboxHandler{
		cfg:    cfg,
		runner: r,
	}
}

// SandboxStatusResponse represents the status of all sandboxes
type SandboxStatusResponse struct {
	Postgres  SandboxDetail `json:"postgres"`
	MySQL     SandboxDetail `json:"mysql"`
	SQLServer SandboxDetail `json:"sqlserver"`
}

// SandboxDetail represents individual sandbox status
type SandboxDetail struct {
	Connected bool   `json:"connected"`
	Error     string `json:"error,omitempty"`
	URI       string `json:"uri,omitempty"`
}

// SandboxTestRequest represents a test query request
type SandboxTestRequest struct {
	DBType string `json:"db_type" binding:"required,oneof=postgresql mysql sqlserver"`
	Query  string `json:"query" binding:"required"`
}

// SandboxTestResponse represents the result of a test query
type SandboxTestResponse struct {
	DBType      string              `json:"db_type"`
	Query       string              `json:"query"`
	Success     bool                `json:"success"`
	Result      *runner.QueryResult `json:"result,omitempty"`
	Error       string              `json:"error,omitempty"`
	ExecutionMs int64               `json:"execution_ms"`
}

// GetStatus godoc
// @Summary     Get sandbox database status
// @Description Check connection status of all sandbox databases
// @Tags        Admin
// @Produce     json
// @Success     200 {object} SandboxStatusResponse
// @Router      /admin/sandbox/status [get]
func (h *SandboxHandler) GetStatus(c *gin.Context) {
	// Check admin role
	role, _ := middlewares.GetUserRole(c)
	if role != "admin" && role != "lecturer" {
		response.Forbidden(c, "Admin or lecturer access required")
		return
	}

	status := SandboxStatusResponse{}
	ctx := context.Background()

	// Test PostgreSQL
	pgResult, err := h.runner.Execute(ctx, runner.DBTypePostgreSQL, "SELECT 1 as test")
	status.Postgres = SandboxDetail{
		Connected: err == nil && pgResult.Error == "",
		Error:     "",
		URI:       maskURI(h.cfg.SandboxPostgresURI),
	}
	if err != nil {
		status.Postgres.Error = err.Error()
	} else if pgResult.Error != "" {
		status.Postgres.Error = pgResult.Error
	}

	// Test MySQL
	mysqlResult, err := h.runner.Execute(ctx, runner.DBTypeMySQL, "SELECT 1 as test")
	status.MySQL = SandboxDetail{
		Connected: err == nil && mysqlResult.Error == "",
		Error:     "",
		URI:       maskURI(h.cfg.SandboxMySQLURI),
	}
	if err != nil {
		status.MySQL.Error = err.Error()
	} else if mysqlResult.Error != "" {
		status.MySQL.Error = mysqlResult.Error
	}

	// Test SQLServer
	mssqlResult, err := h.runner.Execute(ctx, runner.DBTypeSQLServer, "SELECT 1 as test")
	status.SQLServer = SandboxDetail{
		Connected: err == nil && mssqlResult.Error == "",
		Error:     "",
		URI:       maskURI(h.cfg.SandboxSQLServerURI),
	}
	if err != nil {
		status.SQLServer.Error = err.Error()
	} else if mssqlResult.Error != "" {
		status.SQLServer.Error = mssqlResult.Error
	}

	c.JSON(http.StatusOK, status)
}

// TestQuery godoc
// @Summary     Test execute query on sandbox
// @Description Execute a test query on specified sandbox database
// @Tags        Admin
// @Accept      json
// @Produce     json
// @Param       request body SandboxTestRequest true "Test query request"
// @Success     200 {object} SandboxTestResponse
// @Router      /admin/sandbox/test [post]
func (h *SandboxHandler) TestQuery(c *gin.Context) {
	// Check admin role
	role, _ := middlewares.GetUserRole(c)
	if role != "admin" && role != "lecturer" {
		response.Forbidden(c, "Admin or lecturer access required")
		return
	}

	var req SandboxTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	dbType := runner.DBType(req.DBType)
	ctx := context.Background()

	result, err := h.runner.Execute(ctx, dbType, req.Query)

	resp := SandboxTestResponse{
		DBType: req.DBType,
		Query:  req.Query,
	}

	if err != nil {
		resp.Success = false
		resp.Error = err.Error()
	} else if result.Error != "" {
		resp.Success = false
		resp.Error = result.Error
	} else {
		resp.Success = true
		resp.Result = result
		resp.ExecutionMs = result.ExecutionMs
	}

	c.JSON(http.StatusOK, resp)
}

// maskURI masks sensitive information in database URI
func maskURI(uri string) string {
	if uri == "" {
		return "not configured"
	}
	// Simple masking - show only the prefix
	if len(uri) > 20 {
		return uri[:20] + "..."
	}
	return uri
}

// RegisterSandboxRoutes registers sandbox routes
func RegisterSandboxRoutes(router *gin.RouterGroup, handler *SandboxHandler) {
	sandbox := router.Group("/sandbox")
	{
		sandbox.GET("/status", handler.GetStatus)
		sandbox.POST("/test", handler.TestQuery)
	}
}
