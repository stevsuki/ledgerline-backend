-- Any grant on the menu goes with it: role_menu_permissions references menus
-- ON DELETE CASCADE, so this is one statement rather than two.
DELETE FROM menus WHERE id = 'b0000000-0000-0000-0000-000000000015';

UPDATE menus
   SET sort_order = sort_order - 1
 WHERE parent_id = 'a0000000-0000-0000-0000-000000000002'
   AND sort_order >= 4;
