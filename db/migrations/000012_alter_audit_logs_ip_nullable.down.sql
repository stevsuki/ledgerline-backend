DROP INDEX IF EXISTS audit_logs_severity_idx;

UPDATE audit_logs SET ip_address = '0.0.0.0' WHERE ip_address IS NULL;
ALTER TABLE audit_logs ALTER COLUMN ip_address SET NOT NULL;
