package middleware

import (
	"net/http"

	"github.com/bingfengfeifei/kafka-map-go/internal/dto"
	"github.com/bingfengfeifei/kafka-map-go/internal/util"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates authentication token
func AuthMiddleware(tokenCache *util.TokenCache) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Skip authentication for login endpoint
		if ctx.Request.URL.Path == "/api/login" {
			ctx.Next()
			return
		}

		// Get token from header
		token := ctx.GetHeader("X-Auth-Token")
		if token == "" {
			ctx.JSON(http.StatusUnauthorized, dto.Response{
				Code:    http.StatusUnauthorized,
				Message: "Missing authentication token",
			})
			ctx.Abort()
			return
		}

		// Validate token
		username, found := tokenCache.Get(token)
		if !found {
			ctx.JSON(http.StatusUnauthorized, dto.Response{
				Code:    http.StatusUnauthorized,
				Message: "Invalid or expired token",
			})
			ctx.Abort()
			return
		}

		// Store username in context
		ctx.Set("username", username)
		ctx.Next()
	}
}
