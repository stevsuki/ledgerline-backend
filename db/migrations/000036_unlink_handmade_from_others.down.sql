-- 000028's reading, restored: a category belonging to no master belongs to Others.
UPDATE categories
   SET master_category_id = '00000000-0000-0000-0000-000000000007'
 WHERE master_category_id IS NULL;
