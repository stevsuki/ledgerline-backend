CREATE TABLE IF NOT EXISTS master_categories (
    id       UUID         PRIMARY KEY,
    name     VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(100) NOT NULL DEFAULT 'system'
);

CREATE UNIQUE INDEX IF NOT EXISTS master_categories_name_unique_idx
    ON master_categories (LOWER(name));


INSERT INTO master_categories (id, name) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Housing'),
    ('00000000-0000-0000-0000-000000000002', 'Food & Drink'),
    ('00000000-0000-0000-0000-000000000003', 'Transport'),
    ('00000000-0000-0000-0000-000000000004', 'Subscription'),
    ('00000000-0000-0000-0000-000000000005', 'Utilities'),
    ('00000000-0000-0000-0000-000000000006', 'Health');