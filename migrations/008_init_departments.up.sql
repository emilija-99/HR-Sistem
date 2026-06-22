CREATE TABLE IF NOT EXISTS departments (
    id BIGINT PRIMARY KEY NOT NULL,
    name VARCHAR(30) NOT NULL UNIQUE,
    description TEXT,
    status STATUS DEFAULT 'ACTIVE'
);

INSERT INTO departments (id, name, description) VALUES
(1, 'Backend Engineering', 'Server-side development and APIs'),
(2, 'Frontend Engineering', 'Client-side applications and UI'),
(3, 'DevOps', 'CI/CD, infrastructure and cloud operations'),
(4, 'Database Administration', 'Database design, tuning and maintenance'),
(5, 'Human Resources', 'People operations and recruitment'),
(6, 'IT Support', 'Internal IT and technical support'),
(7, 'Engineering Management', 'Team leadership and delivery oversight')
ON CONFLICT (id) DO NOTHING;
