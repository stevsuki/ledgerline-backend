-- How a budget reads its spend: one user's expense rows in one category, within a cycle.
CREATE INDEX IF NOT EXISTS transactions_user_category_occurred_idx
    ON transactions (user_id, category_id, occurred_at) WHERE deleted_at IS NULL;
