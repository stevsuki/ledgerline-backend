ALTER TABLE categories
    DROP COLUMN created_by,
    DROP COLUMN updated_by,
    DROP COLUMN deleted_by;

ALTER TABLE roles
    DROP COLUMN created_by,
    DROP COLUMN updated_by,
    DROP COLUMN deleted_by;

ALTER TABLE users
    DROP COLUMN created_by,
    DROP COLUMN updated_by,
    DROP COLUMN deleted_by;
