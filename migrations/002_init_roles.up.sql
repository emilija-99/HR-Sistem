CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL
);
INSERT INTO roles (name) VALUES ('employee');
INSERT INTO roles (name) VALUES ('superadmin');
INSERT INTO roles (name) VALUES ('admin');
