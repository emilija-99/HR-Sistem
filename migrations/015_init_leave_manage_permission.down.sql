DELETE FROM role_permissions
WHERE permission_id = (SELECT id FROM permissions WHERE code = 'leave.manage');

DELETE FROM permissions WHERE code = 'leave.manage';
