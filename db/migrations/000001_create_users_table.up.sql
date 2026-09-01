CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY,
    email         VARCHAR(255) NOT NULL,
    full_name     VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(20)  NOT NULL DEFAULT 'user',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ,
    CONSTRAINT users_role_check CHECK (role IN ('admin', 'user'))
);

-- Email is unique only across rows that are not soft-deleted.
CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique_idx
    ON users (email) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS users_created_at_idx ON users (created_at DESC);
