package router

import (
	"github.com/gin-gonic/gin"

	"ms-pln-auth/internal/delivery/http/handler"
	"ms-pln-auth/internal/delivery/http/middleware"
	"ms-pln-auth/internal/usecase"
)

// Config holds router configuration
type Config struct {
	AuthHandler    *handler.AuthHandler
	AuthUsecase    usecase.AuthUsecase
	AuthMiddleware *middleware.AuthMiddleware
}

// NewRouter creates and configures the Gin router
func NewRouter(cfg Config) *gin.Engine {
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			// Public routes
			auth.POST("/login", cfg.AuthHandler.Login)
			auth.POST("/register", cfg.AuthHandler.Register)

			// Protected routes
			protected := auth.Group("")
			protected.Use(cfg.AuthMiddleware.Authenticate())
			{
				protected.POST("/logout", cfg.AuthHandler.Logout)
				protected.GET("/me", cfg.AuthHandler.Me)
			}
		}
	}

	return router
}
