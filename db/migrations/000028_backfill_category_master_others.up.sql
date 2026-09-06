-- Rows created before the service defaulted the link; a category with none belongs to Others.
UPDATE categories
SET master_category_id = '00000000-0000-0000-0000-000000000007'
WHERE master_category_id IS NULL;
