package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/stevensuki/ledgerline-backend/pkg/logger"
)

// Logger: structured access log + injects a request_id-tagged logger into the context.
func Logger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if raw := c.Request.URL.RawQuery; raw != "" {
			path = path + "?" + raw
		}

		reqLogger := log.With(
			slog.String("request_id", GetRequestID(c)),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
		)
		c.Request = c.Request.WithContext(logger.WithContext(c.Request.Context(), reqLogger))

		c.Next()

		attrs := []any{
			slog.Int("status", c.Writer.Status()),
			slog.String("query", path),
			slog.String("ip", c.ClientIP()),
			slog.Duration("latency", time.Since(start)),
			slog.Int("bytes", c.Writer.Size()),
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("error", c.Errors.String()))
		}

		switch {
		case c.Writer.Status() >= 500:
			reqLogger.Error("request completed", attrs...)
		case c.Writer.Status() >= 400:
			reqLogger.Warn("request completed", attrs...)
		default:
			reqLogger.Info("request completed", attrs...)
		}
	}
}
