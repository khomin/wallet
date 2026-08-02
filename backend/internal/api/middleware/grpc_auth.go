package middleware

import (
	"context"
	"strings"
	"tracker/internal/core/domain"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type TokenVerifier interface {
	Verify(ctx context.Context, rawToken string) (domain.JwtClaims, error)
}

func UnaryAuthInterceptor(verifier TokenVerifier) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// 1. Skip auth for public endpoints (e.g. Health, Public Coins)
		if isPublicEndpoint(info.FullMethod) {
			return handler(ctx, req)
		}

		// 2. Extract metadata from gRPC context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing request metadata")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization header is required")
		}

		tokenStr := strings.TrimPrefix(authHeader[0], "Bearer ")

		// 3. Verify JWT against Keycloak
		claims, err := verifier.Verify(ctx, tokenStr)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid or expired token: %v", err)
		}

		// 4. Attach user claims/ID to context for downstream handlers
		newCtx := context.WithValue(ctx, "user_id", claims.Subject)

		return handler(newCtx, req)
	}
}

func isPublicEndpoint(method string) bool {
	// Exclude public endpoints matching gRPC method names
	publicMethods := map[string]bool{
		"/price.v1.PriceService/GetCoins":  true,
		"/price.v1.PriceService/GetCoin":   true,
		"/price.v1.PriceService/GetPrices": true,
	}
	return publicMethods[method]
}

// package middleware

// import (
// 	"fmt"
// 	"strings"
// 	"tracker/internal/api/dto"
// 	"tracker/internal/core/domain"

// 	"github.com/coreos/go-oidc/v3/oidc"
// 	"github.com/gin-gonic/gin"
// )

// func AuthGrpc(verifier *oidc.IDTokenVerifier) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		authHeader := c.GetHeader("Authorization")
// 		if authHeader == "" {
// 			dto.UnauthorizedErrorMessage(c, "Authorization header required")
// 			return
// 		}
// 		parts := strings.Split(authHeader, " ") // usually "Bearer <token>"
// 		if len(parts) != 2 || parts[0] != "Bearer" {
// 			dto.UnauthorizedErrorMessage(c, "Authorization header format must be Bearer {token}")
// 			return
// 		}
// 		rawToken := parts[1]

// 		idToken, err := verifier.Verify(c.Request.Context(), rawToken)
// 		if err != nil {
// 			dto.UnauthorizedErrorMessage(c, fmt.Sprintf("Invalid token: %v", err))
// 			return
// 		}
// 		var claims domain.JwtClaims
// 		if err := idToken.Claims(&claims); err != nil {
// 			dto.UnauthorizedErrorMessage(c, "Failed to parse token claims")
// 			return
// 		}
// 		SetOAUTH(c, &claims)
// 		c.Next()
// 	}
// }
