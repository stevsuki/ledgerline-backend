package middleware

import "github.com/gin-gonic/gin"

// SecureHeaders: baseline security headers.
func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-XSS-Protection", "0")
		c.Next()
	}
}
