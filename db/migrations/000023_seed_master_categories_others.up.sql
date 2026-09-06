-- The bucket a spend falls into when none of the named ones fit, and the row
-- the UI hangs "name your own category" on.
INSERT INTO master_categories (id, name) VALUES
    ('00000000-0000-0000-0000-000000000007', 'Others');

-- Users registered before this row existed would never see it: SeedDefaults only
-- runs once, at registration, and nothing revisits it afterwards.
--
-- ON CONFLICT covers anyone who already made a category of their own by this
-- name; the partial unique index on (user_id, LOWER(name)) is what catches it.
INSERT INTO categories (id, user_id, master_category_id, name, type, created_by, updated_by)
SELECT gen_random_uuid(),
       u.id,
       '00000000-0000-0000-0000-000000000007',
       'Others',
       'expense',
       u.id,
       u.id
FROM users u
WHERE u.deleted_at IS NULL
ON CONFLICT DO NOTHING;
