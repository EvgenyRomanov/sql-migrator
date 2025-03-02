-- +gomigrator Up
INSERT INTO test (test, column_int, column_datetime) VALUES ('some test text', 1, '2025-03-02 20:00:00.000001');
INSERT INTO test (test, column_int, column_datetime) VALUES ('some test text 2', 2, '2025-03-02 21:00:00.000001');

-- +gomigrator Down
TRUNCATE test;
