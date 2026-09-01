-- Role becomes a row in roles instead of a fixed string on the user.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS role_id UUID REFERENCES roles (id) ON DELETE RESTRICT;

UPDATE users
SET role_id = CASE role
    WHEN 'admin' THEN '00000000-0000-0000-0000-000000000001'::UUID
    ELSE             '00000000-0000-0000-0000-000000000002'::UUID
END
WHERE role_id IS NULL;

ALTER TABLE users ALTER COLUMN role_id SET NOT NULL;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users DROP COLUMN IF EXISTS role;

CREATE INDEX IF NOT EXISTS users_role_id_idx
    ON users (role_id) WHERE deleted_at IS NULL;
