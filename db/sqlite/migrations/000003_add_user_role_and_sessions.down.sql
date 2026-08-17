-- 000003_add_user_role_and_sessions.down.sql
-- SQLite version

-- Refuse to remove the role if user data now depends on it. This check runs
-- before dropping sessions so a rejected rollback leaves the v0.2.5 data intact.
CREATE TABLE sessions_downgrade_check (
    is_valid INTEGER NOT NULL CHECK (is_valid = 1)
);

INSERT INTO sessions_downgrade_check (is_valid)
SELECT CASE
    WHEN NOT EXISTS (
        SELECT 1
        FROM users
        WHERE user_role_id = (SELECT id FROM roles WHERE role_name = 'user')
    )
    THEN 1
    ELSE 0
END;

DROP TABLE sessions_downgrade_check;

DROP INDEX idx_sessions_expires_at;
DROP TABLE sessions;
DELETE FROM roles WHERE role_name = 'user';
