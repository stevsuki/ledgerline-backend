CREATE TABLE IF NOT EXISTS menus (
    id         UUID PRIMARY KEY,
    parent_id  UUID         REFERENCES menus (id) ON DELETE RESTRICT,
    code       VARCHAR(100) NOT NULL,
    name       VARCHAR(100) NOT NULL,
    path       VARCHAR(255),
    icon       VARCHAR(50),
    sort_order INT          NOT NULL DEFAULT 0,
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT menus_parent_not_self_check CHECK (parent_id IS NULL OR parent_id <> id)
);

-- Code is the stable identifier permission checks rely on, not the display name.
CREATE UNIQUE INDEX IF NOT EXISTS menus_code_unique_idx ON menus (code);

-- Matches the sidebar query: children grouped by parent, in display order.
CREATE INDEX IF NOT EXISTS menus_parent_sort_idx ON menus (parent_id, sort_order);
