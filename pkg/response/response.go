// Package response: a uniform JSON shape for every endpoint.
package response

import "github.com/gin-gonic/gin"

type Success struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"success"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

// Meta: pagination info.
type Meta struct {
	Page       int `json:"page" example:"1"`
	PerPage    int `json:"per_page" example:"10"`
	TotalItems int `json:"total_items" example:"42"`
	TotalPages int `json:"total_pages" example:"5"`
}

// HeaderRequestID: the tracing header; the error body repeats it so one screenshot finds the log.
const HeaderRequestID = "X-Request-ID"

type Error struct {
	Success   bool         `json:"success" example:"false"`
	Message   string       `json:"message" example:"something went wrong"`
	Code      string       `json:"code,omitempty" example:"USER_NOT_FOUND"`
	Errors    []FieldError `json:"errors,omitempty"`
	RequestID string       `json:"request_id,omitempty" example:"9b2c1f2e-6f0a-4a1e-9c7d-2f8b0a1c3d4e"`
}

// FieldError: per-field error detail.
type FieldError struct {
	Field   string `json:"field" example:"email"`
	Message string `json:"message" example:"email must be a valid email address"`
}

func OK(c *gin.Context, status int, message string, data any) {
	c.JSON(status, Success{Success: true, Message: message, Data: data})
}

// Paginated: a list response + meta.
func Paginated(c *gin.Context, status int, message string, data any, meta Meta) {
	c.JSON(status, Success{Success: true, Message: message, Data: data, Meta: &meta})
}

func Fail(c *gin.Context, status int, code, message string, fieldErrors []FieldError) {
	c.AbortWithStatusJSON(status, Error{
		Success:   false,
		Message:   message,
		Code:      code,
		Errors:    fieldErrors,
		RequestID: c.Writer.Header().Get(HeaderRequestID),
	})
}
