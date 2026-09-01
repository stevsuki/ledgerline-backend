CREATE INDEX IF NOT EXISTS users_created_at_idx ON users (created_at DESC);

DROP INDEX IF EXISTS users_list_idx;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_lowercase_check;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_status_check;

ALTER TABLE users DROP COLUMN IF EXISTS password_changed_at;
ALTER TABLE users DROP COLUMN IF EXISTS status;
