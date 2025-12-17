package middleware

import (
	"net/http"
	"restaurant-booking/internal/repository"
	"restaurant-booking/pkg/jwt"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	jwtManager *jwt.Manager
	userRepo   repository.UserRepository
}

func NewAuthMiddleware(jwtManager *jwt.Manager, userRepo repository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		userRepo:   userRepo,
	}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims, err := m.jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		user, err := m.userRepo.GetByID(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)
		c.Set("user", user)

		c.Next()
	}
}
