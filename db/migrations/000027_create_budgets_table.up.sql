CREATE TABLE IF NOT EXISTS budgets (
    id          UUID NOT NULL PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories (id) ON DELETE RESTRICT,
    currency VARCHAR(3) NOT NULL,
    monthly_limit BIGINT NOT NULL,
    alert_threshold_percent SMALLINT NOT NULL DEFAULT 80,
    is_fixed BOOLEAN NOT NULL DEFAULT FALSE,
    rollover BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users (id) ON DELETE SET NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID REFERENCES users (id) ON DELETE SET NULL,
    deleted_at TIMESTAMPTZ,
    deleted_by UUID REFERENCES users (id) ON DELETE SET NULL,
    CONSTRAINT budgets_currency_check  CHECK (currency IN ('IDR', 'USD', 'SGD')),
    CONSTRAINT budgets_limit_check     CHECK (monthly_limit > 0),
    CONSTRAINT budgets_threshold_check CHECK (alert_threshold_percent BETWEEN 1 AND 100),
    -- A fixed commitment only reports when it passes its whole limit.
    CONSTRAINT budgets_fixed_threshold_check
        CHECK (NOT is_fixed OR alert_threshold_percent = 100)
);

CREATE INDEX IF NOT EXISTS budgets_user_id_idx
    ON budgets (user_id, created_at DESC) WHERE deleted_at IS NULL;

-- Keeps the RESTRICT check off a full scan when a category is deleted.
CREATE INDEX IF NOT EXISTS budgets_category_id_idx ON budgets (category_id);

CREATE UNIQUE INDEX IF NOT EXISTS budgets_user_category_unique_idx
    ON budgets (user_id, category_id) WHERE deleted_at IS NULL;
