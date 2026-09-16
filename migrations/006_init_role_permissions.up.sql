CREATE TABLE IF NOT EXISTS role_permissions (
    role_id INT REFERENCES roles(id),
    permission_id INT REFERENCES permissions(id),
    PRIMARY KEY(role_id, permission_id)
);

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.name = 'PLATFORM_ADMIN';

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.code IN (
  'employees.read','employees.write',
  'absence.read.own','absence.read.all','absence.approve','absence.reject',
  'reports.view'
) WHERE r.name = 'HR_ADMIN';

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.code IN (
  'employees.read','absence.read.own','absence.read.all',
  'absence.approve','absence.reject'
) WHERE r.name = 'MANAGER_PORTAL_ACCESS';

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.code IN ('absence.create','absence.read.own')
WHERE r.name = 'EMPLOYEE';
