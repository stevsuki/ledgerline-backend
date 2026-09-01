package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type AuditLogModel struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	UserFullName string     `gorm:"size:100;not null"`
	RoleName     string     `gorm:"size:50;not null"`
	Action       string     `gorm:"size:255;not null"`
	Details      []byte     `gorm:"type:jsonb;not null"`
	DetailText   string     `gorm:"size:500;not null"`
	IPAddress    *string    `gorm:"column:ip_address;type:inet"` // nullable
	Status       string     `gorm:"size:20;not null"`
	Severity     string     `gorm:"size:50;not null"`
	MenuID       *uuid.UUID `gorm:"type:uuid"` // nullable
	Module       string     `gorm:"size:50;not null"`
	CreatedAt    time.Time
}

func (AuditLogModel) TableName() string { return "audit_logs" }

func (m AuditLogModel) ToDomain() *domain.AuditLog {
	log := &domain.AuditLog{
		ID:           m.ID,
		UserID:       m.UserID,
		UserFullName: m.UserFullName,
		RoleName:     m.RoleName,
		Action:       m.Action,
		DetailText:   m.DetailText,
		IPAddress:    m.IPAddress,
		Status:       domain.AuditStatus(m.Status),
		Severity:     domain.AuditSeverity(m.Severity),
		MenuID:       m.MenuID,
		Module:       domain.AuditModule(m.Module),
		CreatedAt:    m.CreatedAt,
	}
	// Handed back as the JSON it was stored as; no reader here needs the kind.
	if len(m.Details) > 0 {
		log.Details = domain.RawAuditDetail(m.Details)
	}
	return log
}

func AuditLogFromDomain(a *domain.AuditLog) (AuditLogModel, error) {
	// Marshalling here is what keeps invalid JSON out of the jsonb column.
	details := []byte("{}")
	if a.Details != nil {
		var err error
		if details, err = json.Marshal(a.Details); err != nil {
			return AuditLogModel{}, fmt.Errorf("encode audit log details: %w", err)
		}
	}

	return AuditLogModel{
		ID:           a.ID,
		UserID:       a.UserID,
		UserFullName: a.UserFullName,
		RoleName:     a.RoleName,
		Action:       a.Action,
		DetailText:   a.DetailText,
		Details:      details,
		IPAddress:    a.IPAddress,
		Status:       string(a.Status),
		Severity:     string(a.Severity),
		MenuID:       a.MenuID,
		Module:       string(a.Module),
		CreatedAt:    a.CreatedAt,
	}, nil
}

func AuditLogsToDomain(models []AuditLogModel) []domain.AuditLog {
	out := make([]domain.AuditLog, 0, len(models))
	for _, m := range models {
		out = append(out, *m.ToDomain())
	}
	return out
}
