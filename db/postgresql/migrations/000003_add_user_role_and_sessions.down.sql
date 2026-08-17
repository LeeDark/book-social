-- 000003_add_user_role_and_sessions.down.sql
-- PostgreSQL version

BEGIN;

-- Refuse to remove the role if user data now depends on it. This check runs
-- before dropping sessions so a rejected rollback leaves the v0.2.5 data intact.
CREATE TEMPORARY TABLE sessions_downgrade_check (
    is_valid BOOLEAN NOT NULL CHECK (is_valid)
) ON COMMIT DROP;

INSERT INTO sessions_downgrade_check (is_valid)
SELECT NOT EXISTS (
    SELECT 1
    FROM users
    WHERE user_role_id = (SELECT id FROM roles WHERE role_name = 'user')
);

DROP TABLE sessions_downgrade_check;

DROP INDEX idx_sessions_expires_at;
DROP TABLE sessions;
DELETE FROM roles WHERE role_name = 'user';

COMMIT;
