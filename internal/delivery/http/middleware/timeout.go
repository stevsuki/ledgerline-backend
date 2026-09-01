package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout: cancel the request context once the deadline passes.
func Timeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		if ctx.Err() != nil && !c.Writer.Written() {
			c.AbortWithStatus(http.StatusGatewayTimeout)
		}
	}
}
