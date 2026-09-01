CREATE TABLE IF NOT EXISTS roles (
    id          UUID PRIMARY KEY,
    name        VARCHAR(50)  NOT NULL,
    description VARCHAR(255),
    is_system   BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

-- Role name is unique (ignoring soft-deleted rows).
CREATE UNIQUE INDEX IF NOT EXISTS roles_name_unique_idx
    ON roles (LOWER(name)) WHERE deleted_at IS NULL;

-- Fixed ids so every environment references the same built-in roles.
INSERT INTO roles (id, name, description, is_system) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Admin', 'Built-in role with access to every menu', TRUE),
    ('00000000-0000-0000-0000-000000000002', 'User',  'Default role assigned to new accounts',   TRUE)
ON CONFLICT (id) DO NOTHING;
