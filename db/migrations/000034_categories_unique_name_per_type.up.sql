-- Income and expense categories are two lists that never mix: the transaction
-- sheet offers only the half matching the direction being recorded, and a
-- budget can only be measured against a spending one. Uniqueness was coarser
-- than that — one name per account, whichever way it ran — which made "Others"
-- for earnings impossible while "Others" for spending existed.
--
-- Widening the key can only free rows, never reject one: every row unique under
-- (user_id, name) is still unique under (user_id, type, name).
DROP INDEX IF EXISTS categories_user_name_unique_idx;

CREATE UNIQUE INDEX IF NOT EXISTS categories_user_type_name_unique_idx
    ON categories (user_id, type, LOWER(name)) WHERE deleted_at IS NULL;
