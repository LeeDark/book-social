-- SQLite version

-- Refuse data loss or silently revoking an appointed administrator.
CREATE TABLE catalog_operations_downgrade_check (
    is_valid INTEGER NOT NULL CHECK (is_valid = 1)
);

INSERT INTO catalog_operations_downgrade_check (is_valid)
SELECT CASE
    WHEN NOT EXISTS (SELECT 1 FROM catalog_audit_log)
        AND NOT EXISTS (
            SELECT 1
            FROM users u
            JOIN roles r ON r.id = u.user_role_id
            WHERE r.role_name = 'admin'
        )
    THEN 1
    ELSE 0
END;

DROP TABLE catalog_operations_downgrade_check;
DROP TABLE catalog_audit_log;

DELETE FROM roles
WHERE role_name = 'admin';
