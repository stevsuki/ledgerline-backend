-- The seeded copies go first. The foreign key is ON DELETE SET NULL, so dropping
-- only the master row would leave every user holding an "Others" with no link.
DELETE FROM categories WHERE master_category_id = '00000000-0000-0000-0000-000000000007';

DELETE FROM master_categories WHERE id = '00000000-0000-0000-0000-000000000007';
