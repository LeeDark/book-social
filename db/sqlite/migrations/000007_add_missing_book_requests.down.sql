-- SQLite version

-- Refuse rollback rather than silently discarding submitted requests.
CREATE TABLE missing_book_requests_downgrade_check (
    is_valid INTEGER NOT NULL CHECK (is_valid = 1)
);

INSERT INTO missing_book_requests_downgrade_check (is_valid)
SELECT CASE
    WHEN NOT EXISTS (SELECT 1 FROM missing_book_requests) THEN 1
    ELSE 0
END;

DROP TABLE missing_book_requests_downgrade_check;
DROP TABLE missing_book_requests;
