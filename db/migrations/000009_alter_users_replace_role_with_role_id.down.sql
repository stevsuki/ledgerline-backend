ALTER TABLE users
    ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'user';

UPDATE users SET role = 'admin'
WHERE role_id = '00000000-0000-0000-0000-000000000001'::UUID;

ALTER TABLE users
    ADD CONSTRAINT users_role_check CHECK (role IN ('admin', 'user'));

DROP INDEX IF EXISTS users_role_id_idx;

ALTER TABLE users DROP COLUMN IF EXISTS role_id;
