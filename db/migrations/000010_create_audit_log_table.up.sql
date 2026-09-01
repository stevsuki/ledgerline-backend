CREATE TABLE audit_log (
    id          UUID PRIMARY KEY,
    user_id     UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    action      VARCHAR(255) NOT NULL,
    details     JSONB        NOT NULL,
    ip_address  INET         NOT NULL,
    severity    VARCHAR(50)  NOT NULL,
    menu_id     UUID         REFERENCES menus (id) ON DELETE SET NULL,
    module      VARCHAR(50)  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS audit_log_created_at_idx ON audit_log (created_at DESC);

CREATE INDEX IF NOT EXISTS audit_log_user_id_idx ON audit_log (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS audit_log_module_idx ON audit_log (module, created_at DESC);