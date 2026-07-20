package middleware

import (
	"fmt"
	"strings"
	"tracker/internal/api/dto"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
)

func Auth(verifier *oidc.IDTokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			dto.UnauthorizedErrorMessage(c, "Authorization header required")
			return
		}
		parts := strings.Split(authHeader, " ") // usually "Bearer <token>"
		if len(parts) != 2 || parts[0] != "Bearer" {
			dto.UnauthorizedErrorMessage(c, "Authorization header format must be Bearer {token}")
			return
		}
		rawToken := parts[1]

		idToken, err := verifier.Verify(c.Request.Context(), rawToken)
		if err != nil {
			dto.UnauthorizedErrorMessage(c, fmt.Sprintf("Invalid token: %v", err))
			return
		}
		var claims map[string]interface{}
		if err := idToken.Claims(&claims); err != nil {
			dto.UnauthorizedErrorMessage(c, "Failed to parse token claims")
			return
		}
		c.Set("claims", claims)
		c.Set("user_id", claims["sub"])
		c.Set("email", claims["email"])
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
