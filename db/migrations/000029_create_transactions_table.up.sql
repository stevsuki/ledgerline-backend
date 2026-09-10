CREATE TABLE IF NOT EXISTS transactions (
    id          UUID         NOT NULL PRIMARY KEY,
    user_id     UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    category_id UUID         NOT NULL REFERENCES categories (id) ON DELETE RESTRICT,
    wallet_id   UUID         NOT NULL REFERENCES wallets (id) ON DELETE RESTRICT,
    name        VARCHAR(150) NOT NULL,
    note        VARCHAR(255),
    type        VARCHAR(20)  NOT NULL,
    amount      BIGINT       NOT NULL,
    currency    VARCHAR(3)   NOT NULL,
    occurred_at TIMESTAMPTZ  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_by  UUID REFERENCES users (id) ON DELETE SET NULL,
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_by  UUID REFERENCES users (id) ON DELETE SET NULL,
    deleted_at  TIMESTAMPTZ,
    deleted_by  UUID REFERENCES users (id) ON DELETE SET NULL,
    CONSTRAINT transactions_type_check     CHECK (type IN ('income', 'expense')),
    CONSTRAINT transactions_amount_check   CHECK ((type = 'income' AND amount > 0) OR (type = 'expense' AND amount < 0)),
    CONSTRAINT transactions_currency_check CHECK (currency IN ('IDR', 'USD', 'SGD'))
);

-- How the ledger screen reads: one user, newest first, paginated.
CREATE INDEX IF NOT EXISTS transactions_user_occurred_idx
    ON transactions (user_id, occurred_at DESC) WHERE deleted_at IS NULL;

-- Keeps the RESTRICT check off a full scan when a category or wallet is deleted.
CREATE INDEX IF NOT EXISTS transactions_category_id_idx ON transactions (category_id);
CREATE INDEX IF NOT EXISTS transactions_wallet_id_idx   ON transactions (wallet_id);
