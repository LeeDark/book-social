-- PostgreSQL version

BEGIN;

-- Prevent a submission racing the emptiness check and being lost on DROP.
LOCK TABLE missing_book_requests IN ACCESS EXCLUSIVE MODE;

-- Refuse rollback rather than silently discarding submitted requests.
CREATE TEMPORARY TABLE missing_book_requests_downgrade_check (
    is_valid BOOLEAN NOT NULL CHECK (is_valid)
) ON COMMIT DROP;

INSERT INTO missing_book_requests_downgrade_check (is_valid)
SELECT NOT EXISTS (SELECT 1 FROM missing_book_requests);

DROP TABLE missing_book_requests_downgrade_check;
DROP TABLE missing_book_requests;

COMMIT;
