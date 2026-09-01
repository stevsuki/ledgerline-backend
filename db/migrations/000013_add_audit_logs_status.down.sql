UPDATE audit_logs SET action = action || '_failed' WHERE status = 'failed';

DROP INDEX IF EXISTS audit_logs_status_idx;
ALTER TABLE audit_logs DROP CONSTRAINT IF EXISTS audit_logs_status_check;
ALTER TABLE audit_logs DROP COLUMN status;
