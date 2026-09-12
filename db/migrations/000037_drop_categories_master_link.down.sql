-- The column comes back empty: which master row each category derived from is
-- not recoverable once dropped, and guessing by name would relink the very
-- hand-made rows 000036 unlinked.
ALTER TABLE categories
    ADD COLUMN IF NOT EXISTS master_category_id UUID REFERENCES master_categories (id) ON DELETE SET NULL;
