-- A category is drawn with a tile beside its name and a slice of the chart ramp
-- behind it, and both are choices somebody made, so both have to survive a
-- reload. Same VARCHAR(50) as menus.icon, roles.icon and wallets.icon: the value
-- is an icon key from the front end's own sprite, not a path and not markup.
--
-- `color` holds a step of that ramp (`c1` … `c7`). It is stored rather than
-- counted from the row's position because the list is editable: once a category
-- can be added above another one, "the seventh category is c7" stops being true
-- and every chart repaints itself the moment somebody reorders anything.
--
-- NOT NULL DEFAULT '' rather than nullable, unlike roles.icon: existing rows
-- would otherwise come back NULL into a non-pointer string. '' is the column's
-- own way of saying "no icon of its own", which the read path resolves — from
-- the master row it derives from, or from its type.
ALTER TABLE categories
    ADD COLUMN icon  VARCHAR(50) NOT NULL DEFAULT '',
    ADD COLUMN color VARCHAR(10) NOT NULL DEFAULT '';
