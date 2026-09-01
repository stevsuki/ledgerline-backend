-- Rendered once at write time so every client shows the same sentence, and so
-- an entry keeps the wording that was true when it happened. The jsonb column
-- stays the source of truth: this one can always be rebuilt from it.
ALTER TABLE audit_logs ADD COLUMN detail_text VARCHAR(500) NOT NULL DEFAULT '';
