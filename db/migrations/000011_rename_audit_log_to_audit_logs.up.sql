-- Table names are plural everywhere else in this schema.
ALTER TABLE audit_log RENAME TO audit_logs;

ALTER INDEX audit_log_pkey RENAME TO audit_logs_pkey;
ALTER INDEX audit_log_created_at_idx RENAME TO audit_logs_created_at_idx;
ALTER INDEX audit_log_user_id_idx RENAME TO audit_logs_user_id_idx;
ALTER INDEX audit_log_module_idx RENAME TO audit_logs_module_idx;
