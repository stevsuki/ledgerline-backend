ALTER TABLE master_categories DROP CONSTRAINT IF EXISTS master_categories_type_check;

ALTER TABLE master_categories
    ADD CONSTRAINT master_categories_type_check
    CHECK (type IN ('income', 'expense', 'both'));

UPDATE master_categories
   SET type = 'both'
 WHERE id = '00000000-0000-0000-0000-000000000007';
