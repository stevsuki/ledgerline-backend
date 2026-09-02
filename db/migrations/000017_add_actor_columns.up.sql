-- Who touched a row, next to the created_at/updated_at/deleted_at columns that
-- already say when. Only the tables an end user writes get these: menus are
-- seeded, password_reset_tokens are machine-written, and audit_logs already
-- name their actor in user_id.
--
-- Every column is nullable: rows written before this migration have no actor,
-- and neither does anything done without a signed-in user (self-registration).
-- ON DELETE SET NULL is the hard-delete fallback; a soft delete keeps the id.

ALTER TABLE users
    ADD COLUMN created_by UUID REFERENCES users (id) ON DELETE SET NULL,
    ADD COLUMN updated_by UUID REFERENCES users (id) ON DELETE SET NULL,
    ADD COLUMN deleted_by UUID REFERENCES users (id) ON DELETE SET NULL;

ALTER TABLE roles
    ADD COLUMN created_by UUID REFERENCES users (id) ON DELETE SET NULL,
    ADD COLUMN updated_by UUID REFERENCES users (id) ON DELETE SET NULL,
    ADD COLUMN deleted_by UUID REFERENCES users (id) ON DELETE SET NULL;

ALTER TABLE categories
    ADD COLUMN created_by UUID REFERENCES users (id) ON DELETE SET NULL,
    ADD COLUMN updated_by UUID REFERENCES users (id) ON DELETE SET NULL,
    ADD COLUMN deleted_by UUID REFERENCES users (id) ON DELETE SET NULL;

-- Existing rows: a category is written by its owner and nobody else, so its
-- history can be filled in exactly. Users and roles are left NULL because
-- their author is genuinely unknown.
UPDATE categories SET created_by = user_id, updated_by = user_id;
