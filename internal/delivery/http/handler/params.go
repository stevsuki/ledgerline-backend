package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

// parseUUIDParam returns a domain error; the caller passes it to handleError like any other.
func parseUUIDParam(c *gin.Context, key string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(key))
	if err != nil {
		return uuid.Nil, domain.InvalidInput(domain.CodeInvalidParam, key+" must be a valid UUID").
			WithField(key).WithCause(err)
	}
	return id, nil
}
