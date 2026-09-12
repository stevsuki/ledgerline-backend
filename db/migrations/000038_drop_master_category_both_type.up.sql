-- 'both' was invented in 000033 so one "Others" row could seed a category in
-- each direction. Adoption cannot use it: the dropdown now hands the server a
-- master id and asks it to make that category, and an id meaning two directions
-- gives it no way to know which one to make.
--
-- Others goes back to being an ordinary expense template. Accounts seeded before
-- this keep both of the "Others" categories they already own — after 000037 they
-- are simply their own categories, and nothing reads the template for them.
UPDATE master_categories SET type = 'expense' WHERE type = 'both';

ALTER TABLE master_categories DROP CONSTRAINT IF EXISTS master_categories_type_check;

ALTER TABLE master_categories
    ADD CONSTRAINT master_categories_type_check
    CHECK (type IN ('income', 'expense'));
