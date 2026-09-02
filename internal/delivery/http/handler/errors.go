// Package handler: HTTP adapter (parse request, call service, map response).
package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/apierr"
)

// handleError: every failure in a handler leaves through here.
func handleError(c *gin.Context, err error) { apierr.Write(c, err) }

// handleBindError: binding/validation failure -> 422 + per-field detail.
func handleBindError(c *gin.Context, err error) { apierr.WriteBind(c, err) }
