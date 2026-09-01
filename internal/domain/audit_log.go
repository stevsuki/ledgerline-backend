package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AuditStatus: did the action succeed, unlike AuditSeverity which says how much it matters.
type AuditStatus string

const (
	AuditStatusSuccess AuditStatus = "success"
	AuditStatusFailed  AuditStatus = "failed"
)

func (s AuditStatus) Valid() bool {
	return s == AuditStatusSuccess || s == AuditStatusFailed
}

type AuditSeverity string

const (
	AuditSeverityInfo     AuditSeverity = "info"
	AuditSeverityWarning  AuditSeverity = "warning"
	AuditSeverityCritical AuditSeverity = "critical"
)

func (s AuditSeverity) Valid() bool {
	return s == AuditSeverityInfo || s == AuditSeverityWarning || s == AuditSeverityCritical
}

// AuditModule mirrors the module filter in the web UI, matching menus.code where one exists.
type AuditModule string

const (
	AuditModuleAuth         AuditModule = "auth"
	AuditModuleBudgets      AuditModule = "budgets"
	AuditModuleDataExport   AuditModule = "data_export"
	AuditModuleInsights     AuditModule = "insights"
	AuditModuleRecurring    AuditModule = "recurring"
	AuditModuleRoles        AuditModule = "roles"
	AuditModuleGoals        AuditModule = "goals"
	AuditModuleSettings     AuditModule = "settings"
	AuditModuleShared       AuditModule = "shared"
	AuditModuleTransactions AuditModule = "transactions"
	AuditModuleUsers        AuditModule = "users"
	AuditModuleWallets      AuditModule = "wallets"
)

// auditModuleList is in the order the filter dropdown shows them.
var auditModuleList = []AuditModule{
	AuditModuleAuth, AuditModuleBudgets, AuditModuleDataExport,
	AuditModuleInsights, AuditModuleRecurring, AuditModuleRoles,
	AuditModuleGoals, AuditModuleSettings, AuditModuleShared,
	AuditModuleTransactions, AuditModuleUsers, AuditModuleWallets,
}

var auditModules = func() map[AuditModule]struct{} {
	set := make(map[AuditModule]struct{}, len(auditModuleList))
	for _, m := range auditModuleList {
		set[m] = struct{}{}
	}
	return set
}()

func (m AuditModule) Valid() bool {
	_, ok := auditModules[m]
	return ok
}

// The three closed vocabularies, for the filter dropdowns.
func AuditModules() []AuditModule {
	return append([]AuditModule(nil), auditModuleList...)
}

func AuditStatuses() []AuditStatus {
	return []AuditStatus{AuditStatusSuccess, AuditStatusFailed}
}

func AuditSeverities() []AuditSeverity {
	return []AuditSeverity{AuditSeverityInfo, AuditSeverityWarning, AuditSeverityCritical}
}

type AuditLog struct {
	ID     uuid.UUID
	UserID uuid.UUID
	// The actor as they were when the entry was written, not following later renames.
	UserFullName string
	RoleName     string
	Action       string
	Details      AuditDetail
	// DetailText is Details rendered for people; Details stays the source of truth.
	DetailText string
	Status     AuditStatus
	Severity   AuditSeverity
	Module     AuditModule
	// IPAddress is nullable: only the HTTP layer knows it.
	IPAddress *string
	// MenuID is nullable: the database clears it when the menu is deleted.
	MenuID    *uuid.UUID
	CreatedAt time.Time
}

type AuditLogFilter struct {
	Search   string
	Limit    int
	Offset   int
	OrderBy  string // ORDER BY clause, may only be filled via pagination.Sortable
	UserID   uuid.UUID
	Status   AuditStatus
	Severity AuditSeverity
	// From and To bound created_at; To is already the exclusive upper bound.
	From   *time.Time
	To     *time.Time
	Module AuditModule
}

type AuditLogRepository interface {
	List(ctx context.Context, filter AuditLogFilter) ([]AuditLog, int, error)
	Create(ctx context.Context, log *AuditLog) error
	Overview(ctx context.Context, window time.Duration) (AuditLogOverview, error)
	DistinctActors(ctx context.Context) ([]AuditActorOption, error)
	// ListRows is List without the count, for walking a large result in batches.
	ListRows(ctx context.Context, filter AuditLogFilter) ([]AuditLog, error)
}

type AuditLogService interface {
	List(ctx context.Context, filter AuditLogFilter) ([]AuditLog, int, error)
	Overview(ctx context.Context) (AuditLogOverview, error)
	Options(ctx context.Context) (AuditLogOptions, error)
	// Export hands each batch to yield so a large export never sits in memory.
	Export(ctx context.Context, filter AuditLogFilter, yield func([]AuditLog) error) error
}

// AuditLogOverview: the cards above the audit table, every count over the same window.
type AuditLogOverview struct {
	WindowDays int
	// Events and Modules answer "Events, 7 days · Across N modules".
	Events  int
	Modules int
	// Sensitive is everything above info severity.
	Sensitive int
	// These two answer "Failed sign-ins · From N addresses".
	FailedSignIns         int
	FailedSignInAddresses int
	// RetentionDays is policy, not a measurement.
	RetentionDays int
}

// AuditActorOption: one actor dropdown entry, named as the log recorded them.
type AuditActorOption struct {
	UserID   uuid.UUID
	FullName string
	RoleName string
}

// AuditLogOptions: the filter dropdowns; actors are recorded, the rest are fixed.
type AuditLogOptions struct {
	Actors     []AuditActorOption
	Modules    []AuditModule
	Statuses   []AuditStatus
	Severities []AuditSeverity
}
