-- A one-off sync. SeedDefaults only runs at registration, so everyone who
-- signed up before it existed is holding an empty category list that nothing
-- would ever fill in.
--
-- On a fresh database this does nothing: there are no users yet, and every
-- account created from here on is seeded by Register itself.
--
-- ON CONFLICT DO NOTHING makes it safe over accounts that already have some of
-- these; the partial unique index on (user_id, LOWER(name)) is what catches the
-- overlap. Everything is an expense because master_categories is a list of
-- names and says nothing about direction.
INSERT INTO categories (id, user_id, master_category_id, name, type, created_by, updated_by)
SELECT gen_random_uuid(), u.id, m.id, m.name, 'expense', u.id, u.id
FROM users u
CROSS JOIN master_categories m
WHERE u.deleted_at IS NULL
ON CONFLICT DO NOTHING;
