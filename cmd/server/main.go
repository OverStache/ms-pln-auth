package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"ms-pln-auth/internal/delivery/http/handler"
	"ms-pln-auth/internal/delivery/http/middleware"
	"ms-pln-auth/internal/delivery/http/router"
	"ms-pln-auth/internal/infrastructure/config"
	"ms-pln-auth/internal/infrastructure/database"
	"ms-pln-auth/internal/repository/postgres"
	"ms-pln-auth/internal/usecase"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Initialize database connection
	db, err := database.NewPostgresConnection(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to database successfully")

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	sessionRepo := postgres.NewSessionRepository(db)

	// Initialize usecases
	authUsecaseConfig := usecase.AuthUsecaseConfig{
		TokenDuration: cfg.Auth.TokenDuration,
		BcryptCost:    cfg.Auth.BcryptCost,
	}
	authUsecase := usecase.NewAuthUsecase(userRepo, sessionRepo, authUsecaseConfig)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUsecase)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authUsecase)

	// Initialize router
	r := router.NewRouter(router.Config{
		AuthHandler:    authHandler,
		AuthUsecase:    authUsecase,
		AuthMiddleware: authMiddleware,
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
