-- Grants administration of leave balances and leave policies.
-- Approval of absence requests is covered by "absence.approve"/"absence.reject".
INSERT INTO permissions (code) VALUES ('leave.manage');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.code = 'leave.manage'
WHERE r.name IN ('PLATFORM_ADMIN', 'HR_ADMIN');
