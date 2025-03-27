package middleware

import (
	"net/http"

	"github.com/1v4n-ML/finance-tracker-api/config"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	if cfg.ApiToken == "" {
		return func(ctx *gin.Context) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Server configuration error, auth token not set"})
		}
	}
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("x-api-key")
		if token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing api key"})
			return
		}
		if token != cfg.ApiToken {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			return
		}
		ctx.Next()
	}
}
