package middleware

import (
	"net/http"
	"restaurant-booking/internal/domain"

	"github.com/gin-gonic/gin"
)

func RequireRole(roles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		role, ok := userRole.(domain.UserRole)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid role"})
			c.Abort()
			return
		}

		for _, allowedRole := range roles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		c.Abort()
	}
}
