-- Outcome is separate from severity: a successful role change is critical,
-- a failed login is only a warning.
ALTER TABLE audit_logs ADD COLUMN status VARCHAR(20);

-- Existing rows encoded the outcome in the action name; move it to the column
-- so one action key covers both outcomes.
UPDATE audit_logs SET status = 'failed' WHERE action LIKE '%\_failed';
UPDATE audit_logs SET status = 'success' WHERE status IS NULL;
UPDATE audit_logs SET action = left(action, length(action) - 7) WHERE action LIKE '%\_failed';

ALTER TABLE audit_logs ALTER COLUMN status SET NOT NULL;

ALTER TABLE audit_logs ADD CONSTRAINT audit_logs_status_check
    CHECK (status IN ('success', 'failed'));

CREATE INDEX IF NOT EXISTS audit_logs_status_idx ON audit_logs (status, created_at DESC);
