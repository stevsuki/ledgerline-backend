package handler

import (
	"encoding/csv"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/dto"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/pkg/logger"
	"github.com/stevensuki/ledgerline-backend/pkg/pagination"
	"github.com/stevensuki/ledgerline-backend/pkg/response"
)

type AuditLogHandler struct {
	auditLogService domain.AuditLogService
}

func NewAuditLogHandler(auditLogService domain.AuditLogService) *AuditLogHandler {
	return &AuditLogHandler{auditLogService: auditLogService}
}

// List godoc
//
//	@Summary		List audit logs
//	@Description	Activity feed. Newest first unless sort says otherwise.
//	@Tags			audit-logs
//	@Produce		json
//	@Security		BearerAuth
//	@Param			search		query		string	false	"Search by actor, action, detail or module"
//	@Param			sort		query		string	false	"Order: user_full_name, role_name, action, detail_text, status, severity, module, user_id, ip_address, created_at. Prefix - for desc"	default(-created_at)
//	@Param			page		query		int		false	"Page"				default(1)
//	@Param			per_page	query		int		false	"Items per page"	default(10)
//	@Success		200			{object}	response.Success{data=[]dto.AuditLogResponseDTO}
//	@Router			/audit-logs [get]
func (h *AuditLogHandler) List(c *gin.Context) {
	var query dto.ListAuditLogQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		handleBindError(c, err)
		return
	}

	orderBy, err := query.OrderBy()
	if err != nil {
		handleError(c, err)
		return
	}

	filter, err := query.ToFilter()
	if err != nil {
		handleError(c, err)
		return
	}

	params := pagination.Params{Page: query.Page, PerPage: query.PerPage}.Normalize()
	filter.Limit, filter.Offset, filter.OrderBy = params.Limit(), params.Offset(), orderBy

	auditLogs, total, err := h.auditLogService.List(c.Request.Context(), filter)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Paginated(c, http.StatusOK, "success", dto.NewAuditLogResponseDTOs(auditLogs), response.Meta{
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalItems: total,
		TotalPages: pagination.TotalPages(total, params.PerPage),
	})
}

// Overview godoc
//
//	@Summary		Audit log overview
//	@Description	Counters for the cards above the audit table. Every count covers the same window (window_days); retention_days is policy, not a measurement.
//	@Tags			audit-logs
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Success{data=dto.AuditLogOverviewResponseDTO}
//	@Router			/audit-logs/overview [get]
func (h *AuditLogHandler) Overview(c *gin.Context) {
	overview, err := h.auditLogService.Overview(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "success", dto.NewAuditLogOverviewResponseDTO(overview))
}

// Options godoc
//
//	@Summary		Audit log filter options
//	@Description	Values for the actor, module, status and severity dropdowns. Actors reflect who has actually been recorded; the rest are fixed vocabularies.
//	@Tags			audit-logs
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Success{data=dto.AuditLogOptionsResponseDTO}
//	@Router			/audit-logs/options [get]
func (h *AuditLogHandler) Options(c *gin.Context) {
	options, err := h.auditLogService.Options(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	response.OK(c, http.StatusOK, "success", dto.NewAuditLogOptionsResponseDTO(options))
}

// Export godoc
//
//	@Summary		Export audit logs as CSV
//	@Description	Same filters as the list, so the file matches the table it was exported from. from/to are dates (YYYY-MM-DD); to covers the whole day. Rows are streamed, newest first.
//	@Tags			audit-logs
//	@Produce		text/csv
//	@Security		BearerAuth
//	@Param			from		query	string	false	"Start date, inclusive"	example(2026-08-01)
//	@Param			to			query	string	false	"End date, inclusive"	example(2026-08-31)
//	@Param			search		query	string	false	"Search by actor, action, detail or module"
//	@Param			user_id		query	string	false	"Filter by actor id"
//	@Param			status		query	string	false	"Filter by outcome"		Enums(success, failed)
//	@Param			severity	query	string	false	"Filter by severity"	Enums(info, warning, critical)
//	@Param			module		query	string	false	"Filter by module"
//	@Success		200			{file}	string	"CSV file"
//	@Router			/audit-logs/export [get]
func (h *AuditLogHandler) Export(c *gin.Context) {
	var query dto.AuditLogRangeDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		handleBindError(c, err)
		return
	}

	filter, err := query.ToFilter()
	if err != nil {
		handleError(c, err)
		return
	}

	// Headers go out first: once writing starts, an error can no longer be JSON.
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="`+query.ExportFilename()+`"`)
	c.Status(http.StatusOK)

	writer := csv.NewWriter(c.Writer)
	if err := writer.Write(dto.AuditLogCSVHeader); err != nil {
		logger.FromContext(c.Request.Context()).Error("write audit log csv", slog.Any("error", err))
		return
	}

	err = h.auditLogService.Export(c.Request.Context(), filter, func(batch []domain.AuditLog) error {
		for _, entry := range batch {
			if err := writer.Write(dto.AuditLogCSVRow(entry)); err != nil {
				return err
			}
		}
		writer.Flush()
		return writer.Error()
	})
	if err != nil {
		// Already streaming: the client gets a truncated file, the log is the only report.
		logger.FromContext(c.Request.Context()).Error("export audit logs", slog.Any("error", err))
		return
	}
	writer.Flush()
}
