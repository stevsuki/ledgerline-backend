-- The IP is only known at the HTTP layer; entries written without one store NULL.
ALTER TABLE audit_logs ALTER COLUMN ip_address DROP NOT NULL;

-- Matches the severity filter in AuditLogRepository.List.
CREATE INDEX IF NOT EXISTS audit_logs_severity_idx ON audit_logs (severity, created_at DESC);
