package postgres

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type auditLogModel struct {
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

func (auditLogModel) TableName() string { return "audit_logs" }

func (m auditLogModel) toDomain() (*domain.AuditLog, error) {
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
	// Handed back as the JSON it was stored as: nothing here needs to know which
	// kind it is, and decoding per kind would need a registry for no gain.
	if len(m.Details) > 0 {
		log.Details = domain.RawAuditDetail(m.Details)
	}
	return log, nil
}

func auditLogFromDomain(a *domain.AuditLog) (auditLogModel, error) {
	// Marshalling here is what keeps invalid JSON out of the jsonb column.
	details := []byte("{}")
	if a.Details != nil {
		var err error
		if details, err = json.Marshal(a.Details); err != nil {
			return auditLogModel{}, fmt.Errorf("encode audit log details: %w", err)
		}
	}

	return auditLogModel{
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

func auditLogsToDomain(models []auditLogModel) ([]domain.AuditLog, error) {
	out := make([]domain.AuditLog, 0, len(models))
	for _, m := range models {
		log, err := m.toDomain()
		if err != nil {
			return nil, err
		}
		out = append(out, *log)
	}
	return out, nil
}
