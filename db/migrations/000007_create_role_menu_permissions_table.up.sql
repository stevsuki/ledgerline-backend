CREATE TABLE IF NOT EXISTS role_menu_permissions (
    role_id     UUID        NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    menu_id     UUID        NOT NULL REFERENCES menus (id) ON DELETE CASCADE,
    can_create  BOOLEAN     NOT NULL DEFAULT FALSE,
    can_read    BOOLEAN     NOT NULL DEFAULT FALSE,
    can_update  BOOLEAN     NOT NULL DEFAULT FALSE,
    can_delete  BOOLEAN     NOT NULL DEFAULT FALSE,
    can_approve BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, menu_id)
);

-- Reverse lookup: which roles reach a given menu. The PK already covers role_id.
CREATE INDEX IF NOT EXISTS role_menu_permissions_menu_id_idx
    ON role_menu_permissions (menu_id);
