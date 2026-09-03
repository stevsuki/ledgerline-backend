-- The role list is rendered with a badge next to the name, so the icon the
-- front end picked has to survive a reload. Same VARCHAR(50) as menus.icon
-- and wallets.icon: the value is an icon key, not a path.
ALTER TABLE roles
    ADD COLUMN icon VARCHAR(50);
