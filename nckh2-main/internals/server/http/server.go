package http

import (
	"backend/configs"
	"backend/db"
	adminHttp "backend/internals/admin/controller/http"
	authHttp "backend/internals/auth/controller/http"
	examHttp "backend/internals/exam/controller/http"
	problemHttp "backend/internals/problem/controller/http"
	submissionHttp "backend/internals/submission/controller/http"
	topicHttp "backend/internals/topic/controller/http"
	"backend/pkgs/jwt"
	"backend/pkgs/middlewares"
	"backend/pkgs/redis"
	"backend/pkgs/runner"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	engine      *gin.Engine
	cfg         *configs.Config
	database    *db.Database
	cache       redis.IRedis
	jwtProv     jwt.JWTProvider
	queryRunner runner.Runner
}

// NewServer is injectable by DI container
func NewServer(cfg *configs.Config, database *db.Database, cache redis.IRedis, jwtProv jwt.JWTProvider, queryRunner runner.Runner) *Server {
	return &Server{
		engine:      gin.Default(),
		cfg:         cfg,
		database:    database,
		cache:       cache,
		jwtProv:     jwtProv,
		queryRunner: queryRunner,
	}
}

func (s *Server) Run() error {
	_ = s.engine.SetTrustedProxies(nil)

	// Always set release mode to disable debug logs
	gin.SetMode(gin.ReleaseMode)

	// Middlewares
	s.engine.Use(middlewares.RecoveryMiddleware())
	s.engine.Use(middlewares.LoggerMiddleware())
	s.engine.Use(middlewares.CorsMiddleware())

	// Health check
	s.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	s.engine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "SQL Exam System API v1.0",
			"docs":    "/api/v1",
		})
	})

	// Map routes
	s.MapRoutes()

	// Run server
	return s.engine.Run(fmt.Sprintf(":%d", s.cfg.HTTPPort))
}

func (s *Server) MapRoutes() {
	v1 := s.engine.Group("/api/v1")

	// Create auth middleware for protected routes
	authMiddleware := middlewares.AuthMiddleware(s.jwtProv, s.cache)

	// Auth routes (register, login, logout, refresh)
	authHttp.Routes(v1, s.database, s.cfg, s.cache, s.jwtProv)

	// Topic routes (CRUD with role protection)
	topicHttp.Routes(v1, s.database, authMiddleware)

	// Problem routes (CRUD with role protection)
	problemHttp.Routes(v1, s.database, authMiddleware)

	// Submission routes (run, submit, list)
	submissionHttp.Routes(v1, s.database, s.queryRunner, s.cfg, authMiddleware)

	// Exam routes (CRUD, participants, student actions)
	examHttp.Routes(v1, s.database, s.queryRunner, s.cfg, authMiddleware)

	// Admin routes (user import, stats)
	adminHttp.Routes(v1, s.database, authMiddleware)
}
