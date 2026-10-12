CREATE TABLE IF NOT EXISTS recurrings (
    id            UUID NOT NULL PRIMARY KEY,
    user_id       UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    category_id   UUID NOT NULL REFERENCES categories (id) ON DELETE RESTRICT,
    wallet_id     UUID NOT NULL REFERENCES wallets (id) ON DELETE RESTRICT,
    name          VARCHAR(100) NOT NULL,
    amount        BIGINT NOT NULL,
    currency      VARCHAR(3) NOT NULL,
    frequency     VARCHAR(10) NOT NULL,
    next_due_date DATE NOT NULL,
    -- The day of the month the schedule was set on, so a 31st survives February.
    anchor_day    SMALLINT NOT NULL,
    is_paused     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users (id) ON DELETE SET NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID REFERENCES users (id) ON DELETE SET NULL,
    deleted_at TIMESTAMPTZ,
    deleted_by UUID REFERENCES users (id) ON DELETE SET NULL,
    CONSTRAINT recurrings_currency_check   CHECK (currency IN ('IDR', 'USD', 'SGD')),
    CONSTRAINT recurrings_amount_check     CHECK (amount > 0),
    CONSTRAINT recurrings_frequency_check  CHECK (frequency IN ('weekly', 'monthly', 'quarterly', 'annual')),
    CONSTRAINT recurrings_anchor_day_check CHECK (anchor_day BETWEEN 1 AND 31)
);

CREATE INDEX IF NOT EXISTS recurrings_user_due_idx
    ON recurrings (user_id, next_due_date) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS recurrings_category_id_idx ON recurrings (category_id);
CREATE INDEX IF NOT EXISTS recurrings_wallet_id_idx ON recurrings (wallet_id);
