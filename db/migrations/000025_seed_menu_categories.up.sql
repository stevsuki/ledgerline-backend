-- The categories screen, as a menu of its own.
--
-- It sits under `money` beside Budgets rather than inside it: a budget is a
-- limit somebody sets on a category, so the two are read together, but only one
-- of them is a real resource here. `/categories` has full CRUD and its own
-- error codes; budgets has no table and no endpoint at all, and the only trace
-- of it in this schema is `slug=budget` on the category options call — which is
-- the budget form *reading* this list.
--
-- Keeping them as two menus is also what keeps the permission separable:
-- role_menu_permissions is keyed per menu, and renaming or removing a category
-- reaches every transaction filed under it, which setting a limit does not.
UPDATE menus
   SET sort_order = sort_order + 1
 WHERE parent_id = 'a0000000-0000-0000-0000-000000000002'
   AND sort_order >= 3;

INSERT INTO menus (id, parent_id, code, name, path, icon, sort_order) VALUES
    ('b0000000-0000-0000-0000-000000000015', 'a0000000-0000-0000-0000-000000000002', 'categories', 'Categories', NULL, 'tag', 3)
ON CONFLICT (id) DO NOTHING;
