-- Rebuilds the table exactly as 000002 and 000017 left it.
CREATE TABLE IF NOT EXISTS categories (
    id         UUID         PRIMARY KEY,
    user_id    UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name       VARCHAR(100) NOT NULL,
    type       VARCHAR(20)  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    created_by UUID REFERENCES users (id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users (id) ON DELETE SET NULL,
    deleted_by UUID REFERENCES users (id) ON DELETE SET NULL,
    CONSTRAINT categories_type_check CHECK (type IN ('income', 'expense'))
);

CREATE UNIQUE INDEX IF NOT EXISTS categories_user_name_unique_idx
    ON categories (user_id, LOWER(name)) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS categories_user_id_idx ON categories (user_id, created_at DESC);
