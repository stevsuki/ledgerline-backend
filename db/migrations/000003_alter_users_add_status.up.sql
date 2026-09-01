-- Account state controlled by an admin, not a measure of user activity.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'enabled';

-- Lets a password reset invalidate tokens issued before the change.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS password_changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE users
    ADD CONSTRAINT users_status_check
    CHECK (status IN ('enabled', 'disabled', 'invited'));

-- Normalise existing rows before the constraint can reject them.
UPDATE users SET email = LOWER(email) WHERE email <> LOWER(email);

-- The unique index sits on the raw column, so lowercase must be enforced here.
ALTER TABLE users
    ADD CONSTRAINT users_email_lowercase_check
    CHECK (email = LOWER(email));

-- Matches the list query: WHERE deleted_at IS NULL ORDER BY created_at DESC, id ASC.
CREATE INDEX IF NOT EXISTS users_list_idx
    ON users (created_at DESC, id ASC) WHERE deleted_at IS NULL;

DROP INDEX IF EXISTS users_created_at_idx;
