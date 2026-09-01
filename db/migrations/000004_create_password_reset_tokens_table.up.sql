CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id         UUID PRIMARY KEY,
    user_id    UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    otp_hash   VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ  NOT NULL,
    attempts   SMALLINT     NOT NULL DEFAULT 0,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT password_reset_tokens_attempts_check CHECK (attempts >= 0)
);

-- Lookup path when verifying an OTP: newest unused token for one user.
CREATE INDEX IF NOT EXISTS password_reset_tokens_user_idx
    ON password_reset_tokens (user_id, created_at DESC) WHERE used_at IS NULL;

-- Supports the periodic purge of expired rows.
CREATE INDEX IF NOT EXISTS password_reset_tokens_expires_idx
    ON password_reset_tokens (expires_at);
