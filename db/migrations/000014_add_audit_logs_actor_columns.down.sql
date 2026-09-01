DROP INDEX IF EXISTS audit_logs_user_full_name_idx;
ALTER TABLE audit_logs DROP COLUMN role_name;
ALTER TABLE audit_logs DROP COLUMN user_full_name;
