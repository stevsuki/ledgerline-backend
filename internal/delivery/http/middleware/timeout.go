package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/apierr"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

// Timeout: cancel the request context once the deadline passes.
func Timeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		// Last resort: the handler returned without writing anything.
		if ctx.Err() != nil && !c.Writer.Written() {
			apierr.Write(c, domain.ErrTimeout)
		}
	}
}
