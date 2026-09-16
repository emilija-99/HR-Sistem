CREATE TABLE IF NOT EXISTS absence_types (
    id INT PRIMARY KEY,
    type_name VARCHAR(40) NOT NULL,
    code VARCHAR(40) NOT NULL UNIQUE,
    is_paid BOOLEAN NOT NULL,
    status STATUS NOT NULL DEFAULT 'ACTIVE'
);

INSERT INTO absence_types (id, type_name, code, is_paid) VALUES
(1, 'Vacation',        'VACATION',    TRUE),
(2, 'Parental Leave',  'PARENTAL',    TRUE),
(3, 'Sick Leave',      'SICK',        TRUE),
(4, 'Training Leave',  'TRAINING',    TRUE),
(5, 'Disability Leave','DISABILITY',  TRUE),
(6, 'Personal Leave',  'PERSONAL',    TRUE)
ON CONFLICT (id) DO NOTHING;
