-- Denormalised so an entry keeps the name and role the actor had at the time,
-- and stays readable after a rename, a role change, or a deleted user.
ALTER TABLE audit_logs ADD COLUMN user_full_name VARCHAR(100) NOT NULL DEFAULT '';
ALTER TABLE audit_logs ADD COLUMN role_name      VARCHAR(50)  NOT NULL DEFAULT '';

UPDATE audit_logs a
SET user_full_name = u.full_name,
    role_name      = COALESCE(r.name, '')
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE u.id = a.user_id;

CREATE INDEX IF NOT EXISTS audit_logs_user_full_name_idx ON audit_logs (user_full_name);
