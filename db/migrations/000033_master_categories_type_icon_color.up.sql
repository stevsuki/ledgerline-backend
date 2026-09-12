-- The seed template gains the three things it was silently deciding for the
-- front end: which way a row runs, and what it is drawn with.
--
-- `type` is read only by SeedDefaults. It is a default, not a lock: nothing
-- stops a person filing their own income category under the Housing row.
-- 'both' is for a row that is a real bucket on either side — "Others" catches
-- unnamed spending and unnamed earnings alike, and stays one row for both.
--
-- `icon` and `color` move here from a hardcoded map in the web UI, so adding a
-- master row no longer means editing TypeScript to give it a tile.
ALTER TABLE master_categories
    ADD COLUMN type  VARCHAR(20) NOT NULL DEFAULT 'expense',
    ADD COLUMN icon  VARCHAR(50) NOT NULL DEFAULT '',
    ADD COLUMN color VARCHAR(10) NOT NULL DEFAULT '';

ALTER TABLE master_categories
    ADD CONSTRAINT master_categories_type_check
    CHECK (type IN ('income', 'expense', 'both'));

-- What the web UI's MASTER_DEFAULTS held, now stated where the rows live.
UPDATE master_categories SET icon = 'home',    color = 'c1' WHERE id = '00000000-0000-0000-0000-000000000001';
UPDATE master_categories SET icon = 'cup',     color = 'c2' WHERE id = '00000000-0000-0000-0000-000000000002';
UPDATE master_categories SET icon = 'car',     color = 'c3' WHERE id = '00000000-0000-0000-0000-000000000003';
UPDATE master_categories SET icon = 'monitor', color = 'c4' WHERE id = '00000000-0000-0000-0000-000000000004';
UPDATE master_categories SET icon = 'wifi',    color = 'c5' WHERE id = '00000000-0000-0000-0000-000000000005';
UPDATE master_categories SET icon = 'heart',   color = 'c6' WHERE id = '00000000-0000-0000-0000-000000000006';

UPDATE master_categories
   SET type = 'both', icon = 'bag', color = 'c7'
 WHERE id = '00000000-0000-0000-0000-000000000007';

-- Four earning buckets, kept deliberately fewer than the spending ones: a
-- person has many ways to spend and few to earn, and a seeded row nobody ever
-- files anything under is noise in every picker it appears in.
--
-- Bonus is apart from Salary because it is irregular — a THR folded into Salary
-- makes one month spike and every month-on-month delta meaningless. Investment
-- is apart from both because it is money that did not come from working, which
-- is the distinction that makes a savings rate worth reading.
INSERT INTO master_categories (id, name, type, icon, color) VALUES
    ('00000000-0000-0000-0000-000000000008', 'Salary',      'income', 'bank',  'c1'),
    ('00000000-0000-0000-0000-000000000009', 'Bonus',       'income', 'gift',  'c2'),
    ('00000000-0000-0000-0000-00000000000a', 'Side income', 'income', 'cash',  'c3'),
    ('00000000-0000-0000-0000-00000000000b', 'Investment',  'income', 'chart', 'c4')
ON CONFLICT DO NOTHING;
