package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"ms-pln-auth/internal/usecase"
)

// AuthMiddleware handles authentication middleware
type AuthMiddleware struct {
	authUsecase usecase.AuthUsecase
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(authUsecase usecase.AuthUsecase) *AuthMiddleware {
	return &AuthMiddleware{
		authUsecase: authUsecase,
	}
}

// Authenticate returns a Gin middleware that validates the bearer token
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "No token provided",
			})
			return
		}

		user, err := m.authUsecase.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid or expired token",
			})
			return
		}

		// Set user and token in context for handlers to use
		c.Set("user", user)
		c.Set("token", token)
		c.Set("user_id", user.ID)

		c.Next()
	}
}

// extractBearerToken extracts the bearer token from the Authorization header
func extractBearerToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}
