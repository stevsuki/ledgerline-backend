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

type Error struct {
	Success bool         `json:"success" example:"false"`
	Message string       `json:"message" example:"something went wrong"`
	Code    string       `json:"code,omitempty" example:"NOT_FOUND"`
	Errors  []FieldError `json:"errors,omitempty"`
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
		Success: false,
		Message: message,
		Code:    code,
		Errors:  fieldErrors,
	})
}
