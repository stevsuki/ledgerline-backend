-- Only the untouched ones go back: a seeded income category that already
-- carries transactions is somebody's record, not this migration's to remove.
DELETE FROM categories c
 WHERE c.type = 'income'
   AND c.master_category_id IN (
       '00000000-0000-0000-0000-000000000007',
       '00000000-0000-0000-0000-000000000008',
       '00000000-0000-0000-0000-000000000009',
       '00000000-0000-0000-0000-00000000000a',
       '00000000-0000-0000-0000-00000000000b'
   )
   AND NOT EXISTS (SELECT 1 FROM transactions t WHERE t.category_id = c.id)
   AND NOT EXISTS (SELECT 1 FROM budgets b WHERE b.category_id = c.id);

-- The icon fill is left in place: which rows were blank before it ran is no
-- longer knowable, and a tile is harmless where a deleted row would not be.
