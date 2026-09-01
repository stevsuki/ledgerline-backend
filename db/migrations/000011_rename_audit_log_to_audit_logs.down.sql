ALTER INDEX audit_logs_module_idx RENAME TO audit_log_module_idx;
ALTER INDEX audit_logs_user_id_idx RENAME TO audit_log_user_id_idx;
ALTER INDEX audit_logs_created_at_idx RENAME TO audit_log_created_at_idx;
ALTER INDEX audit_logs_pkey RENAME TO audit_log_pkey;

ALTER TABLE audit_logs RENAME TO audit_log;
