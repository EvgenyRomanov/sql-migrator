-- +gomigrator Up
ALTER TABLE test
    ADD COLUMN column_int int,
    ADD COLUMN column_datetime timestamp;

-- +gomigrator Down
ALTER TABLE test
    DROP COLUMN column_int,
    DROP COLUMN column_datetime;
