// Package middleware: cross-cutting concerns for the HTTP layer.
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/pkg/response"
)

const (
	// Shared with the error envelope, which repeats the id in the body.
	HeaderRequestID  = response.HeaderRequestID
	ContextRequestID = "request_id"
)

// RequestID: a unique ID per request, for tracing.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(HeaderRequestID)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set(ContextRequestID, requestID)
		c.Writer.Header().Set(HeaderRequestID, requestID)
		c.Next()
	}
}

// GetRequestID from the gin context.
func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(ContextRequestID); ok {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}
