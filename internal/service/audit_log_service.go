package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/pkg/logger"
)

type auditLogService struct {
	auditLogRepo domain.AuditLogRepository
}

func NewAuditLogService(auditLogRepo domain.AuditLogRepository) domain.AuditLogService {
	return &auditLogService{auditLogRepo: auditLogRepo}
}

func (s *auditLogService) List(ctx context.Context, filter domain.AuditLogFilter) ([]domain.AuditLog, int, error) {
	return s.auditLogRepo.List(ctx, filter)
}

func recordAudit(ctx context.Context, repo domain.AuditLogRepository, entry *domain.AuditLog) {
	id, err := uuid.NewV7()
	if err == nil {
		entry.ID = id
		// Rendered here so no call site can forget it.
		if entry.Details != nil {
			entry.DetailText = entry.Details.Text()
		}
		err = repo.Create(ctx, entry)
	}
	if err != nil {
		logger.FromContext(ctx).Error("write audit log",
			slog.String("action", entry.Action), slog.Any("error", err))
	}
}

const (
	// The cards above the audit table all describe the same window.
	auditOverviewWindow = 7 * 24 * time.Hour
	// Policy only: nothing deletes old rows yet.
	auditRetentionDays = 365
)

func (s *auditLogService) Overview(ctx context.Context) (domain.AuditLogOverview, error) {
	overview, err := s.auditLogRepo.Overview(ctx, auditOverviewWindow)
	if err != nil {
		return domain.AuditLogOverview{}, err
	}

	overview.WindowDays = int(auditOverviewWindow.Hours() / 24)
	overview.RetentionDays = auditRetentionDays
	return overview, nil
}

func (s *auditLogService) Options(ctx context.Context) (domain.AuditLogOptions, error) {
	actors, err := s.auditLogRepo.DistinctActors(ctx)
	if err != nil {
		return domain.AuditLogOptions{}, err
	}

	return domain.AuditLogOptions{
		Actors:     actors,
		Modules:    domain.AuditModules(),
		Statuses:   domain.AuditStatuses(),
		Severities: domain.AuditSeverities(),
	}, nil
}

// auditExportBatch keeps a large export off the heap: rows are handed to the
// writer batch by batch instead of collected first.
const auditExportBatch = 500

func (s *auditLogService) Export(
	ctx context.Context, filter domain.AuditLogFilter, yield func([]domain.AuditLog) error,
) error {
	filter.Limit = auditExportBatch
	filter.Offset = 0

	for {
		rows, err := s.auditLogRepo.ListRows(ctx, filter)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		if err := yield(rows); err != nil {
			return err
		}
		// A short batch means the end; asking again would only cost a query.
		if len(rows) < auditExportBatch {
			return nil
		}
		filter.Offset += auditExportBatch
	}
}
