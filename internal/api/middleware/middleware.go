package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}
		parts := strings.Split(authHeader, " ") // usually "Bearer <token>"
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			return
		}

		// tokenString := parts[1]

		// --- 1. VALIDATE TOKEN HERE ---
		// Use your Auth0/JWT validation logic.
		// For this example, let's assume it validated successfully and returned the "sub" claim:
		userID := "auth0|65c8abc1234def5678" // This would come from token.Claims["sub"]

		// --- 2. INJECT INTO CONTEXT ---
		// Set the userID in Gin's context so any downstream handler can grab it
		c.Set("userID", userID)

		c.Next()
	}
}

func UserIDFromContext(c *gin.Context) (string, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	userID, ok := value.(string)
	return userID, ok
}
