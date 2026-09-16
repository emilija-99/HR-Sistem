CREATE TABLE IF NOT EXISTS positions (
    id BIGSERIAL PRIMARY KEY,
    department_id BIGINT NOT NULL REFERENCES departments(id),
    title VARCHAR(100) NOT NULL,
    level POSITION_LEVEL NOT NULL DEFAULT 'ROOKIE',
    description TEXT,
    status STATUS NOT NULL DEFAULT 'ACTIVE',
    UNIQUE(department_id, title, level)
);

INSERT INTO positions (department_id, title, level, description) VALUES
(1,'Backend Engineer','ROOKIE','Entry backend developer'),
(1,'Backend Engineer','JUNIOR','Junior backend developer'),
(1,'Backend Engineer','MEDIOR','Mid-level backend developer'),
(1,'Backend Engineer','SENIOR','Senior backend developer'),
(1,'Backend Engineer','LEAD','Leads backend team');

INSERT INTO positions (department_id, title, level, description) VALUES
(2,'Frontend Engineer','ROOKIE','Entry frontend developer'),
(2,'Frontend Engineer','JUNIOR','Junior frontend developer'),
(2,'Frontend Engineer','MEDIOR','Mid-level frontend developer'),
(2,'Frontend Engineer','SENIOR','Senior frontend developer'),
(2,'Frontend Engineer','LEAD','Leads frontend team');

INSERT INTO positions (department_id, title, level, description) VALUES
(3,'DevOps Engineer','JUNIOR','CI/CD and automation'),
(3,'DevOps Engineer','MEDIOR','Infrastructure and monitoring'),
(3,'DevOps Engineer','SENIOR','Cloud architecture'),
(3,'DevOps Engineer','LEAD','DevOps team lead');

INSERT INTO positions (department_id, title, level, description) VALUES
(4,'Database Administrator','JUNIOR','DB maintenance'),
(4,'Database Administrator','MEDIOR','Query optimization and tuning'),
(4,'Database Administrator','SENIOR','Database architecture'),
(4,'Database Administrator','LEAD','DBA team lead');

INSERT INTO positions (department_id, title, level, description) VALUES
(5,'HR Specialist','JUNIOR','Recruitment support'),
(5,'HR Specialist','MEDIOR','Employee lifecycle management'),
(5,'HR Manager','SENIOR','HR department leadership');

INSERT INTO positions (department_id, title, level, description) VALUES
(6,'IT Support Specialist','JUNIOR','Helpdesk and support'),
(6,'IT Support Specialist','MEDIOR','Systems administration'),
(6,'IT Manager','SENIOR','IT department leadership');

INSERT INTO positions (department_id, title, level, description) VALUES
(7,'Engineering Manager','LEAD','Leads engineering teams'),
(7,'Head of Engineering','LEAD','Leads entire engineering org');
