package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/pkg/response"
)

// parseUUIDParam writes its own error response when the param is invalid.
func parseUUIDParam(c *gin.Context, key string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(key))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PARAM", key+" must be a valid UUID", nil)
		return uuid.Nil, err
	}
	return id, nil
}
