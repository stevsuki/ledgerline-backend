-- SeedDefaults runs once, at registration, and nothing revisits it — so every
-- account that existed before 000033 holds spending categories only, and cannot
-- record a single rupiah of income: checkCategoryType refuses an income filed
-- under an expense category, and they own nothing else.
INSERT INTO categories (id, user_id, master_category_id, name, type, icon, color, created_by, updated_by)
SELECT gen_random_uuid(), u.id, m.id, m.name, k.kind, m.icon, m.color, u.id, u.id
FROM users u
CROSS JOIN master_categories m
CROSS JOIN LATERAL unnest(
    CASE WHEN m.type = 'both' THEN ARRAY['expense', 'income'] ELSE ARRAY[m.type] END
) AS k(kind)
WHERE u.deleted_at IS NULL
  AND k.kind = 'income'
ON CONFLICT DO NOTHING;

-- The tiles the web UI used to resolve from a hardcoded table, written onto the
-- rows themselves so that table can go. Only blanks are filled: a category
-- somebody already gave a tile or a colour keeps the one they chose.
UPDATE categories c
   SET icon  = CASE WHEN c.icon  = '' THEN m.icon  ELSE c.icon  END,
       color = CASE WHEN c.color = '' THEN m.color ELSE c.color END
  FROM master_categories m
 WHERE c.master_category_id = m.id
   AND c.deleted_at IS NULL
   AND (c.icon = '' OR c.color = '');
