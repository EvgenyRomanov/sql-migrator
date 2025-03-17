-- +gomigrator Up
CREATE TABLE IF NOT EXISTS test (
    id serial NOT NULL,
    test text
);
SELECT * FROM test;

-- +gomigrator Down
DROP TABLE test;
