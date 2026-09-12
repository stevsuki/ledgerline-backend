-- 000028 gave every category without a master link the "Others" row, on the
-- reading that a category belonging to nothing belongs to the catch-all. But
-- the link is also what marks a row AS the catch-all: it is how the app decides
-- which category the add-transaction sheet offers to name, and which one cannot
-- be deleted. So a hand-made "Education" came back as the bucket for unnamed
-- spending, and with a fallback now existing per direction, whichever such row
-- sorted first would have taken the naming prompt for its whole direction.
--
-- A row is the catch-all only if it still carries the master row's own name.
-- Everything else goes back to having no master, which is what it always was.
UPDATE categories c
   SET master_category_id = NULL
  FROM master_categories m
 WHERE c.master_category_id = m.id
   AND m.id = '00000000-0000-0000-0000-000000000007'
   AND LOWER(c.name) <> LOWER(m.name);
