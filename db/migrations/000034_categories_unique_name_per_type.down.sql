-- Narrowing again can collide: an account holding "Others" both ways has to
-- lose one before the old index can be built.
DROP INDEX IF EXISTS categories_user_type_name_unique_idx;

CREATE UNIQUE INDEX IF NOT EXISTS categories_user_name_unique_idx
    ON categories (user_id, LOWER(name)) WHERE deleted_at IS NULL;
