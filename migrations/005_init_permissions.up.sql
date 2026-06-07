CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    code TEXT UNIQUE NOT NULL
);
INSERT INTO permissions (code) VALUES
('users.manage'),
('roles.manage'),
('employees.read'),
('employees.write'),
('absence.create'),
('absence.read.own'),
('absence.read.all'),
('absence.approve'),
('absence.reject'),
('reports.view');
