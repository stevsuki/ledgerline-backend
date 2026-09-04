-- Drops the per-user categories table. Nothing referenced it and it held no
-- rows; its replacement lands in a later migration, once its shape is settled.
--
-- 000002 and 000017 are left alone: they are already applied, and a fresh
-- database still has to replay them in order before reaching this drop.
DROP TABLE IF EXISTS categories;
