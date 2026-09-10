-- How a wallet reads its running balance: one wallet's rows since it was last stated.
CREATE INDEX IF NOT EXISTS transactions_wallet_occurred_idx
    ON transactions (wallet_id, occurred_at) WHERE deleted_at IS NULL;
