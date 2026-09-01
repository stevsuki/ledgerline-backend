package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/stevensuki/ledgerline-backend/pkg/response"
)

// Recovery: catch panics, stack trace goes to the log only.
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("panic recovered",
					slog.String("request_id", GetRequestID(c)),
					slog.Any("error", err),
					slog.String("stack", string(debug.Stack())),
				)
				response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR",
					"an internal server error occurred", nil)
			}
		}()
		c.Next()
	}
}
