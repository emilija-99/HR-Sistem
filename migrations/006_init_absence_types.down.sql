CREATE TABLE IF NOT EXISTS absence_types (
    id INT PRIMARY KEY,
    type_name VARCHAR(40) NOT NULL,
    code VARCHAR(40) NOT NULL UNIQUE,
    is_paid BOOLEAN NOT NULL,
    status STATUS not null default 'ACTIVE',
    UNIQUE(code)
);
insert
  into
  absence_types
values (
  1,
    'Vacation',
    'VACATION',
  true
    );

INSERT INTO absence_types  VALUES(
  2,
  'Parental Leave',
  'PARENTAL',
  true);

INSERT INTO absence_types VALUES(
  3,
  'Sick Leave',
  'SICK',
  true);
INSERT INTO absence_types VALUES(
  4,
  'Training Leave',
  'TRAINING',
  true);

insert
  into
  absence_types
values(5, 'Disability Leave', 'DISABILITY', true)

INSERT INTO absence_types VALUES(
  6,
  'Personal Leave',
  'PERSONAL',
  true);
