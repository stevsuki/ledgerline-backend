package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/stevensuki/ledgerline-backend/pkg/response"
)

// NotFound: handler for unregistered routes.
func NotFound(c *gin.Context) {
	response.Fail(c, http.StatusNotFound, "ROUTE_NOT_FOUND", "endpoint not found", nil)
}

// MethodNotAllowed: handler for disallowed methods.
func MethodNotAllowed(c *gin.Context) {
	response.Fail(c, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", nil)
}
