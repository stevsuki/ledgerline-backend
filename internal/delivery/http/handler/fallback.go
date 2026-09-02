package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/pkg/response"
)

// NotFound: handler for unregistered routes.
func NotFound(c *gin.Context) {
	response.Fail(c, http.StatusNotFound, domain.CodeRouteNotFound, "endpoint not found", nil)
}

// MethodNotAllowed: handler for disallowed methods.
func MethodNotAllowed(c *gin.Context) {
	response.Fail(c, http.StatusMethodNotAllowed, domain.CodeMethodNotAllowed, "method not allowed", nil)
}
