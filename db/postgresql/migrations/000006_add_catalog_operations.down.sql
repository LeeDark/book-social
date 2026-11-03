-- PostgreSQL version

BEGIN;

-- Refuse data loss or silently revoking an appointed administrator.
CREATE TEMPORARY TABLE catalog_operations_downgrade_check (
    is_valid BOOLEAN NOT NULL CHECK (is_valid)
) ON COMMIT DROP;

INSERT INTO catalog_operations_downgrade_check (is_valid)
SELECT NOT EXISTS (SELECT 1 FROM catalog_audit_log)
    AND NOT EXISTS (
        SELECT 1
        FROM users u
        JOIN roles r ON r.id = u.user_role_id
        WHERE r.role_name = 'admin'
    );

DROP TABLE catalog_operations_downgrade_check;
DROP TABLE catalog_audit_log;

DELETE FROM roles
WHERE role_name = 'admin';

COMMIT;
