package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/pkg/response"
)

type HealthHandler struct {
	db      *gorm.DB
	version string
}

func NewHealthHandler(db *gorm.DB, version string) *HealthHandler {
	return &HealthHandler{db: db, version: version}
}

// Liveness godoc
//
//	@Summary	Liveness probe
//	@Tags		health
//	@Produce	json
//	@Success	200	{object}	response.Success
//	@Router		/health [get]
func (h *HealthHandler) Liveness(c *gin.Context) {
	response.OK(c, http.StatusOK, "ok", gin.H{"status": "healthy", "version": h.version})
}

// Readiness godoc
//
//	@Summary		Readiness probe
//	@Description	Checks the database connection
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	response.Success
//	@Failure		503	{object}	response.Error
//	@Router			/health/ready [get]
func (h *HealthHandler) Readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	sqlDB, err := h.db.DB()
	if err == nil {
		err = sqlDB.PingContext(ctx)
	}
	if err != nil {
		response.Fail(c, http.StatusServiceUnavailable, "DB_UNAVAILABLE", "the database is unreachable", nil)
		return
	}
	response.OK(c, http.StatusOK, "ok", gin.H{"database": "up"})
}
