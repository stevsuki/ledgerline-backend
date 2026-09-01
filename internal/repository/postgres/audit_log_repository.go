package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/internal/repository/postgres/model"
)

const defaultAuditLogOrder = "created_at DESC, id ASC"

type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) domain.AuditLogRepository {
	return &auditLogRepository{db: db}
}

// filtered applies every condition, shared so the list and the export cannot drift apart.
func (r *auditLogRepository) filtered(ctx context.Context, filter domain.AuditLogFilter) *gorm.DB {
	query := dbFrom(ctx, r.db).Model(&model.AuditLogModel{})

	if filter.Search != "" {
		// Searches detail_text, not the raw jsonb whose key names would match blindly.
		keyword := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where(
			`(LOWER(user_full_name) LIKE ? OR LOWER(action) LIKE ?
			  OR LOWER(detail_text) LIKE ? OR LOWER(module) LIKE ?)`,
			keyword, keyword, keyword, keyword,
		)
	}
	if filter.UserID != uuid.Nil {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", string(filter.Status))
	}
	if filter.Severity != "" {
		query = query.Where("severity = ?", string(filter.Severity))
	}
	if filter.Module != "" {
		query = query.Where("module = ?", string(filter.Module))
	}
	if filter.From != nil {
		query = query.Where("created_at >= ?", *filter.From)
	}
	if filter.To != nil {
		query = query.Where("created_at < ?", *filter.To)
	}
	return query
}

// ListRows is List without the count, for walking a large result in batches.
func (r *auditLogRepository) ListRows(ctx context.Context, filter domain.AuditLogFilter) ([]domain.AuditLog, error) {
	orderBy := filter.OrderBy
	if orderBy == "" {
		orderBy = defaultAuditLogOrder
	}

	var logs []model.AuditLogModel
	err := r.filtered(ctx, filter).
		Order(orderBy).Limit(filter.Limit).Offset(filter.Offset).
		Find(&logs).Error
	if err != nil {
		return nil, wrapErr("list audit logs", err)
	}
	return model.AuditLogsToDomain(logs), nil
}

func (r *auditLogRepository) List(ctx context.Context, filter domain.AuditLogFilter) ([]domain.AuditLog, int, error) {
	query := r.filtered(ctx, filter)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, wrapErr("count audit logs", err)
	}
	if total == 0 {
		return []domain.AuditLog{}, 0, nil
	}

	orderBy := filter.OrderBy
	if orderBy == "" {
		orderBy = defaultAuditLogOrder
	}

	var logs []model.AuditLogModel
	err := query.Order(orderBy).Limit(filter.Limit).Offset(filter.Offset).Find(&logs).Error
	if err != nil {
		return nil, 0, wrapErr("list audit logs", err)
	}
	return model.AuditLogsToDomain(logs), int(total), nil
}

func (r *auditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	row, err := model.AuditLogFromDomain(log)
	if err != nil {
		return err
	}
	if err := dbFrom(ctx, r.db).Create(&row).Error; err != nil {
		return wrapErr("create audit log", err)
	}

	log.CreatedAt = row.CreatedAt
	return nil
}

// Overview counts every card in one pass, so the numbers describe the same rows.
func (r *auditLogRepository) Overview(ctx context.Context, window time.Duration) (domain.AuditLogOverview, error) {
	const q = `
		SELECT
			count(*)                                       AS events,
			count(DISTINCT module)                         AS modules,
			count(*) FILTER (WHERE severity <> 'info')     AS sensitive,
			count(*) FILTER (WHERE failed_sign_in)         AS failed_sign_ins,
			count(DISTINCT ip_address) FILTER (WHERE failed_sign_in)
			                                               AS failed_sign_in_addresses
		FROM (
			SELECT module, severity, ip_address,
			       action = 'auth.login' AND status = 'failed' AS failed_sign_in
			FROM audit_logs
			WHERE created_at >= now() - CAST(? AS interval)
		) recent`

	var overview domain.AuditLogOverview
	interval := fmt.Sprintf("%d seconds", int(window.Seconds()))
	if err := dbFrom(ctx, r.db).Raw(q, interval).Scan(&overview).Error; err != nil {
		return domain.AuditLogOverview{}, wrapErr("audit log overview", err)
	}
	return overview, nil
}

// DistinctActors: one row per person with their latest name, ordered by name.
func (r *auditLogRepository) DistinctActors(ctx context.Context) ([]domain.AuditActorOption, error) {
	// Aliased to full_name: the scan matches on field name, not column name.
	const q = `
		SELECT user_id, user_full_name AS full_name, role_name FROM (
			SELECT DISTINCT ON (user_id) user_id, user_full_name, role_name, created_at
			FROM audit_logs
			ORDER BY user_id, created_at DESC
		) actors
		ORDER BY user_full_name`

	var rows []domain.AuditActorOption
	if err := dbFrom(ctx, r.db).Raw(q).Scan(&rows).Error; err != nil {
		return nil, wrapErr("list audit log actors", err)
	}
	return rows, nil
}
