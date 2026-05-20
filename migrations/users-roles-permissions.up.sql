CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    is_active boolean default TRUE,
    created_at TIMESTAMP DEFAULT now(), -- created now
    created_by BIGINT NULL REFERENCES users(id), -- created by NULL meaning that user is creator of his user account
    updated_at TIMESTAMP, -- update can be null
    updated_by BIGINT NULL REFERENCES users(id)
);

SELECT * FROM users;

INSERT INTO users (email, password_hash, is_active, created_by, updated_at, updated_by)
VALUES ('hr-admin@hr-sistem.com', 'testTest!', TRUE, NULL, NULL, NULL);
INSERT INTO users (email, password_hash, is_active, created_by, updated_at, updated_by)
VALUES ('hr-assistent@hr-sistem.com', 'testTest!', TRUE, NULL, NULL, NULL);

INSERT INTO users (email, password_hash, is_active, created_by, updated_at, updated_by)
VALUES ('milica.nikolic@hr-sistem.com', 'testTest!', TRUE, 1, NULL, NULL);
INSERT INTO users (email, password_hash, is_active, created_by, updated_at, updated_by)
VALUES ('nikola.stojanovic@hr-sistem.com', 'testTest!', TRUE, 1, NULL, NULL);
INSERT INTO users (email, password_hash, is_active, created_by, updated_at, updated_by)
VALUES ('jelena.pavlovic@hr-sistem.com', 'testTest!', TRUE, 1, NULL, NULL);
INSERT INTO users (email, password_hash, is_active, created_by, updated_at, updated_by)
VALUES ('stefan.markovic@hr-sistem.com', 'testTest!', TRUE, 1, NULL, NULL);
INSERT INTO users (email, password_hash, is_active, created_by, updated_at, updated_by)
VALUES ('ivana.ivanovic@hr-sistem.com', 'testTest!', TRUE, 1, NULL, NULL);
INSERT INTO users (email, password_hash, is_active, created_by, updated_at, updated_by)
VALUES ('luka.djordjevic@hr-sistem.com', 'testTest!', TRUE, 2, NULL, NULL);
INSERT INTO users (email, password_hash, is_active, created_by, updated_at, updated_by)
VALUES ('teodora.milenkovic@hr-sistem.com', 'testTest!',TRUE , 2, NULL, NULL);
INSERT INTO users (email, password_hash, is_active, created_by, updated_at, updated_by)
VALUES ('nemanja.ristic@hr-sistem.com', 'testTest!', TRUE, 2, NULL, NULL);
INSERT INTO users (email, password_hash, is_active, created_by, updated_at, updated_by)
VALUES ('ana.petrovic@hr-sistem.com', 'testTest!', TRUE, 2, NULL, NULL);
INSERT INTO users (email, password_hash, is_active, created_by, updated_at, updated_by)
VALUES ('proba.test@hr-sistem.com', 'testTest!', TRUE, 2, NULL, NULL);

-- usage to reset serial to last column insert id
SELECT setval(
    pg_get_serial_sequence('users', 'id'),
    COALESCE((SELECT MAX(id) FROM users), 0) + 1,
    false
);

-- System roles
--CREATE TYPE SYSTEM_ROLES AS ENUM ('PLATFORM_ADMIN','HR_ADMIN','EMPLOYEE','MANAGER_PORTAL_ACCESS');
CREATE TABLE roles(
  id SERIAL PRIMARY KEY,
  name TEXT UNIQUE NOT NULL
);
INSERT INTO roles (name) VALUES
('PLATFORM_ADMIN'),
('HR_ADMIN'),
('EMPLOYEE'),
('MANAGER_PORTAL_ACCESS');

-- System permissions
CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    code TEXT UNIQUE NOT NULL
);
INSERT INTO permissions (code) VALUES
-- platform
('users.manage'),
('roles.manage'),

-- employees
('employees.read'),
('employees.write'),

-- absence requests
('absence.create'),
('absence.read.own'),
('absence.read.all'),
('absence.approve'),
('absence.reject'),

-- reports
('reports.view');


-- Role permissions
CREATE TABLE role_permissions (
    role_id INT REFERENCES roles(id),
    permission_id INT REFERENCES permissions(id),
    PRIMARY KEY(role_id, permission_id)
);
-- User with role `PLATFORM_ADMIN` has all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r CROSS JOIN permissions p
WHERE r.name = 'PLATFORM_ADMIN';

-- User with a role `HR_ADMIN` has operational access
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.code IN (
  'employees.read',
  'employees.write',
  'absence.read',
  'absence.approve',
  'absence.reject',
  'reports.view'
)
WHERE r.name = 'HR_ADMIN';

-- User with role `MANAGER_PORTAL_ACCESS` has approval only
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.code IN (
  'employees.read',
  'absence.read',
  'absence.approve',
  'absence.reject'
)
WHERE r.name = 'MANAGER_PORTAL_ACCESS'

-- User with role `EMPLOYEE` has self-service only
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.code IN (
  'absence.create',
  'absence.read'
)
WHERE r.name = 'EMPLOYEE';

CREATE TABLE user_roles (
    user_id BIGINT REFERENCES users(id),
    role_id INT REFERENCES roles(id),
    PRIMARY KEY(user_id, role_id)
);

-- PLATFORM_ADMIN
INSERT INTO user_roles VALUES (1, (SELECT id FROM roles WHERE name='PLATFORM_ADMIN'));

-- HR_ADMIN
INSERT INTO user_roles VALUES (2, (SELECT id FROM roles WHERE name='HR_ADMIN'));

-- MANAGER_PORTAL_ACCESS
INSERT INTO user_roles VALUES (3, (SELECT id FROM roles WHERE name='MANAGER_PORTAL_ACCESS'));
INSERT INTO user_roles VALUES (8, (SELECT id FROM roles WHERE name='MANAGER_PORTAL_ACCESS'));
--INSERT INTO user_roles VALUES (, (SELECT id FROM roles WHERE name='MANAGER_PORTAL_ACCESS'));

-- EMPLOYEE
INSERT INTO user_roles VALUES (4, (SELECT id FROM roles WHERE name='EMPLOYEE'));
INSERT INTO user_roles VALUES (5, (SELECT id FROM roles WHERE name='EMPLOYEE'));
INSERT INTO user_roles VALUES (6, (SELECT id FROM roles WHERE name='EMPLOYEE'));
INSERT INTO user_roles VALUES (7, (SELECT id FROM roles WHERE name='EMPLOYEE'));
--INSERT INTO user_roles VALUES (7, (SELECT id FROM roles WHERE name='EMPLOYEE'));

-- QUERY: get users with their roles
SELECT
    ur.user_id,
    u.email,
    ur.role_id,
    r.name AS role_name
FROM user_roles ur
JOIN users u ON u.id = ur.user_id
JOIN roles r ON r.id = ur.role_id
WHERE u.is_active = TRUE
ORDER BY u.email, r.name;

-- QUERY: get users with roles and persmissions
SELECT
    u.id AS user_id,
    u.email,
    r.name AS role_name,
    p.code AS permission_code
FROM users u
JOIN user_roles ur ON ur.user_id = u.id
JOIN roles r ON r.id = ur.role_id
JOIN role_permissions rp ON rp.role_id = r.id
JOIN permissions p ON p.id = rp.permission_id
WHERE u.is_active = TRUE
ORDER BY u.email, r.name, p.code;
