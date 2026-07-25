package middleware

import (
	"tracker/internal/core/domain"

	"github.com/gin-gonic/gin"
)

const (
	OAUTH            = "oauth_claims"
	DeviceAuthorized = "X-Device-Authorized"
)

func GetOAUTH(c *gin.Context) (*domain.JwtClaims, bool) {
	v, exists := c.Get(OAUTH)
	if !exists {
		return nil, false
	}
	v2, ok := v.(domain.JwtClaims)
	return &v2, ok
}

func SetOAUTH(c *gin.Context, claims *domain.JwtClaims) {
	c.Set(OAUTH, *claims)
}
