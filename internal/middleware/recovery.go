package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"splitobillo/pkg/apperr"
	"splitobillo/pkg/response"
)

func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					"request_id", c.GetString(CtxRequestID),
					"panic", r,
					"path", c.Request.URL.Path,
				)
				if !c.Writer.Written() {
					response.Error(c, apperr.Internal("Internal server error"))
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}
