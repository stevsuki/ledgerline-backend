CREATE TABLE IF NOT EXISTS wallets (
    id       UUID         PRIMARY KEY,
    user_id  UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name     VARCHAR(100) NOT NULL,
    type     VARCHAR(20)  NOT NULL,
    currency VARCHAR(3)   NOT NULL,
    reference VARCHAR(50),
    icon      VARCHAR(50),
    balance BIGINT NOT NULL DEFAULT 0,
    balance_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    balance_updated_by UUID REFERENCES users (id) ON DELETE SET NULL,
    include_in_total BOOLEAN NOT NULL DEFAULT TRUE,
    credit_limit BIGINT,
    due_day      SMALLINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users (id) ON DELETE SET NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID REFERENCES users (id) ON DELETE SET NULL,
    deleted_at TIMESTAMPTZ,
    deleted_by UUID REFERENCES users (id) ON DELETE SET NULL,
    CONSTRAINT wallets_type_check     CHECK (type IN ('bank', 'ewallet', 'card', 'cash')),
    CONSTRAINT wallets_currency_check CHECK (currency IN ('IDR', 'USD', 'SGD')),
    CONSTRAINT wallets_due_day_check  CHECK (due_day BETWEEN 1 AND 31)
);

CREATE INDEX IF NOT EXISTS wallets_user_id_idx ON wallets (user_id, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS wallets_user_name_unique_idx
    ON wallets (user_id, LOWER(name)) WHERE deleted_at IS NULL;
