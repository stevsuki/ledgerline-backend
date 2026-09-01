package dto

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/pkg/pagination"
)

var auditLogSort = pagination.Sortable{
	Allowed: pagination.Whitelist{
		"user_full_name": "audit_logs.user_full_name",
		"role_name":      "audit_logs.role_name",
		"action":         "audit_logs.action",
		"detail_text":    "audit_logs.detail_text",
		"status":         "audit_logs.status",
		"severity":       "audit_logs.severity",
		"module":         "audit_logs.module",
		"user_id":        "audit_logs.user_id",
		"created_at":     "audit_logs.created_at",
		"ip_address":     "audit_logs.ip_address",
	},
	Default:    "-created_at",
	TieBreaker: "audit_logs.id",
}

type ListAuditLogQuery struct {
	AuditLogRange
	Sort    string `form:"sort" example:"-created_at"`
	Page    int    `form:"page" binding:"omitempty,min=1" example:"1"`
	PerPage int    `form:"per_page" binding:"omitempty,min=1,max=100" example:"10"`
}

func (q ListAuditLogQuery) OrderBy() (string, error) { return auditLogSort.OrderBy(q.Sort) }

type AuditLogResponse struct {
	ID           string `json:"id" example:"01a05b87-ac4c-7332-a78c-c60bcec135de"`
	Action       string `json:"action" example:"auth.login"`
	UserFullName string `json:"user_full_name" example:"Rangga Aditama"`
	RoleName     string `json:"role_name" example:"Owner"`
	DetailText   string `json:"detail_text" example:"Email + password · Chrome on macOS"`
	Status       string `json:"status" enums:"success,failed" example:"success"`
	Severity     string `json:"severity" enums:"info,warning,critical" example:"info"`
	Module       string `json:"module" example:"auth"`
	// Details is one of the kind-tagged shapes in domain; see AuditKind.
	Details domain.AuditDetail `json:"details"`
	UserID  string             `json:"user_id" example:"01a05b87-ac4c-7332-a78c-c60bcec135de"`
	// IPAddress is empty for entries written outside an HTTP request.
	IPAddress string `json:"ip_address" example:"103.28.14.7"`
	CreatedAt string `json:"created_at" example:"2026-08-27 19:41:00"`
}

func NewAuditLogResponse(a domain.AuditLog) AuditLogResponse {
	res := AuditLogResponse{
		ID:           a.ID.String(),
		Action:       a.Action,
		UserFullName: a.UserFullName,
		RoleName:     a.RoleName,
		DetailText:   a.DetailText,
		Status:       string(a.Status),
		Severity:     string(a.Severity),
		Module:       string(a.Module),
		Details:      a.Details,
		UserID:       a.UserID.String(),
		CreatedAt:    a.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if a.IPAddress != nil {
		res.IPAddress = *a.IPAddress
	}
	return res
}

func NewAuditLogResponses(a []domain.AuditLog) []AuditLogResponse {
	var res []AuditLogResponse
	for _, v := range a {
		res = append(res, NewAuditLogResponse(v))
	}
	return res
}

// AuditLogOverviewResponse feeds the four cards above the audit table. Plain
// counts, not rendered cards: the labels, tones and wording stay in the UI.
type AuditLogOverviewResponse struct {
	// WindowDays is the period every count below covers.
	WindowDays int `json:"window_days" example:"7"`
	Events     int `json:"events" example:"28"`
	Modules    int `json:"modules" example:"9"`
	// Sensitive counts everything above info severity.
	Sensitive             int `json:"sensitive" example:"8"`
	FailedSignIns         int `json:"failed_sign_ins" example:"1"`
	FailedSignInAddresses int `json:"failed_sign_in_addresses" example:"1"`
	// RetentionDays is policy; no job deletes old rows yet.
	RetentionDays int `json:"retention_days" example:"365"`
}

func NewAuditLogOverviewResponse(o domain.AuditLogOverview) AuditLogOverviewResponse {
	return AuditLogOverviewResponse{
		WindowDays:            o.WindowDays,
		Events:                o.Events,
		Modules:               o.Modules,
		Sensitive:             o.Sensitive,
		FailedSignIns:         o.FailedSignIns,
		FailedSignInAddresses: o.FailedSignInAddresses,
		RetentionDays:         o.RetentionDays,
	}
}

// FilterOption: one entry of a filter dropdown.
type FilterOption struct {
	Value string `json:"value" example:"auth"`
	Label string `json:"label" example:"Auth"`
}

// ActorOption carries the role so the dropdown can show it beside the name.
type ActorOption struct {
	Value string `json:"value" example:"01a05b87-ac4c-7332-a78c-c60bcec135de"`
	Label string `json:"label" example:"Rangga Aditama"`
	Role  string `json:"role" example:"Owner"`
}

type AuditLogOptionsResponse struct {
	Actors     []ActorOption  `json:"actors"`
	Modules    []FilterOption `json:"modules"`
	Statuses   []FilterOption `json:"statuses"`
	Severities []FilterOption `json:"severities"`
}

// Display labels live here, not in the domain: they are presentation, and this
// is the only place that has to change to translate them.
var auditModuleLabels = map[domain.AuditModule]string{
	domain.AuditModuleAuth:         "Auth",
	domain.AuditModuleBudgets:      "Budgets",
	domain.AuditModuleDataExport:   "Data export",
	domain.AuditModuleInsights:     "Insights & reports",
	domain.AuditModuleRecurring:    "Recurring items",
	domain.AuditModuleRoles:        "Role management",
	domain.AuditModuleGoals:        "Savings goals",
	domain.AuditModuleSettings:     "Settings",
	domain.AuditModuleShared:       "Shared budgets",
	domain.AuditModuleTransactions: "Transactions",
	domain.AuditModuleUsers:        "User management",
	domain.AuditModuleWallets:      "Wallets & accounts",
}

var auditStatusLabels = map[domain.AuditStatus]string{
	domain.AuditStatusSuccess: "Success",
	domain.AuditStatusFailed:  "Failed",
}

var auditSeverityLabels = map[domain.AuditSeverity]string{
	domain.AuditSeverityInfo:     "Info",
	domain.AuditSeverityWarning:  "Warning",
	domain.AuditSeverityCritical: "Critical",
}

func NewAuditLogOptionsResponse(o domain.AuditLogOptions) AuditLogOptionsResponse {
	res := AuditLogOptionsResponse{
		Actors:     make([]ActorOption, 0, len(o.Actors)),
		Modules:    make([]FilterOption, 0, len(o.Modules)),
		Statuses:   make([]FilterOption, 0, len(o.Statuses)),
		Severities: make([]FilterOption, 0, len(o.Severities)),
	}

	for _, a := range o.Actors {
		res.Actors = append(res.Actors, ActorOption{
			Value: a.UserID.String(), Label: a.FullName, Role: a.RoleName,
		})
	}
	for _, m := range o.Modules {
		res.Modules = append(res.Modules, FilterOption{Value: string(m), Label: auditModuleLabels[m]})
	}
	for _, s := range o.Statuses {
		res.Statuses = append(res.Statuses, FilterOption{Value: string(s), Label: auditStatusLabels[s]})
	}
	for _, s := range o.Severities {
		res.Severities = append(res.Severities, FilterOption{Value: string(s), Label: auditSeverityLabels[s]})
	}
	return res
}

// auditDateLayout: the date a picker sends, without a time part.
const auditDateLayout = "2006-01-02"

// AuditLogRange is the shared filter of the list and the export, so a CSV can
// never contain a different set of rows than the table it was exported from.
type AuditLogRange struct {
	Search   string `form:"search" example:"login"`
	UserID   string `form:"user_id"`
	Status   string `form:"status" example:"failed"`
	Severity string `form:"severity" example:"info"`
	Module   string `form:"module" example:"auth"`
	From     string `form:"from" example:"2026-08-01"`
	To       string `form:"to" example:"2026-08-31"`
}

// ToFilter validates the free-text parameters and turns them into a filter.
func (q AuditLogRange) ToFilter() (domain.AuditLogFilter, error) {
	filter := domain.AuditLogFilter{Search: q.Search}

	if q.UserID != "" {
		id, err := uuid.Parse(q.UserID)
		if err != nil {
			return filter, fmt.Errorf("%w: user_id must be a uuid", domain.ErrInvalidInput)
		}
		filter.UserID = id
	}
	if q.Status != "" {
		if status := domain.AuditStatus(q.Status); status.Valid() {
			filter.Status = status
		} else {
			return filter, fmt.Errorf("%w: unknown status %q", domain.ErrInvalidInput, q.Status)
		}
	}
	if q.Severity != "" {
		if severity := domain.AuditSeverity(q.Severity); severity.Valid() {
			filter.Severity = severity
		} else {
			return filter, fmt.Errorf("%w: unknown severity %q", domain.ErrInvalidInput, q.Severity)
		}
	}
	if q.Module != "" {
		if module := domain.AuditModule(q.Module); module.Valid() {
			filter.Module = module
		} else {
			return filter, fmt.Errorf("%w: unknown module %q", domain.ErrInvalidInput, q.Module)
		}
	}

	if q.From != "" {
		from, err := time.Parse(auditDateLayout, q.From)
		if err != nil {
			return filter, fmt.Errorf("%w: from must be YYYY-MM-DD", domain.ErrInvalidInput)
		}
		filter.From = &from
	}
	if q.To != "" {
		to, err := time.Parse(auditDateLayout, q.To)
		if err != nil {
			return filter, fmt.Errorf("%w: to must be YYYY-MM-DD", domain.ErrInvalidInput)
		}
		// The picker means the whole day, so the bound is the next midnight.
		to = to.AddDate(0, 0, 1)
		filter.To = &to
	}
	if filter.From != nil && filter.To != nil && !filter.From.Before(*filter.To) {
		return filter, fmt.Errorf("%w: from must be on or before to", domain.ErrInvalidInput)
	}
	return filter, nil
}

// ExportFilename names the download after the range it covers.
func (q AuditLogRange) ExportFilename() string {
	switch {
	case q.From != "" && q.To != "":
		return fmt.Sprintf("audit-log_%s_%s.csv", q.From, q.To)
	case q.From != "":
		return fmt.Sprintf("audit-log_from-%s.csv", q.From)
	case q.To != "":
		return fmt.Sprintf("audit-log_until-%s.csv", q.To)
	default:
		return "audit-log.csv"
	}
}

// AuditLogCSVHeader names the columns of the export, in order.
var AuditLogCSVHeader = []string{
	"timestamp", "actor", "role", "action", "status",
	"severity", "module", "detail", "ip_address",
}

// AuditLogCSVRow renders one entry in the same order as the header.
func AuditLogCSVRow(a domain.AuditLog) []string {
	ip := ""
	if a.IPAddress != nil {
		ip = *a.IPAddress
	}
	return []string{
		a.CreatedAt.Format(time.RFC3339),
		a.UserFullName,
		a.RoleName,
		a.Action,
		string(a.Status),
		string(a.Severity),
		string(a.Module),
		a.DetailText,
		ip,
	}
}
