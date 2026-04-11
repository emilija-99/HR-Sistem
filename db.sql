CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active boolean default true,
);

CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL
);

INSERT INTO roles (name) VALUES ('user');
INSERT INTO roles (name) VALUES ('admin');
INSERT INTO roles (name) VALUES ('approver');


CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL
);
# me permissions

INSERT into permissions (id, name) values (7, 'user.read');
INSERT into permissions (id, name) values (8, 'user.edit');
INSERT into permissions (id, name) values (9, 'user.delete');

# absence permissions
INSERT into permissions (id, name) values (1, 'absence.read');
INSERT into permissions (id, name) values (2, 'absence.create');
INSERT into permissions (id, name) values (3, 'absence.approve');
INSERT into permissions (id, name) values (4, 'absence.reject');
INSERT into permissions (id, name) values (5, 'absence.delete');
INSERT into permissions (id, name) values (6, 'absence.cancel');



CREATE TABLE IF NOT EXISTS user_roles (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    role_id INT REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id INT REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INT REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

# user can manage (read,create,cancel) only own absences
insert into role_permissions (role_id, permission_id) values (1,1);
insert into role_permissions (role_id, permission_id) values (1,2);
insert into role_permissions (role_id, permission_id) values (1,6);

# admin is superadmin, have all premissions

insert into role_permissions (role_id, permission_id) values (2,1);
insert into role_permissions (role_id, permission_id) values (2,2);
insert into role_permissions (role_id, permission_id) values (2,3);
insert into role_permissions (role_id, permission_id) values (2,4);
insert into role_permissions (role_id, permission_id) values (2,5);
insert into role_permissions (role_id, permission_id) values (2,6);

# approver can be team lead, project manager, superviors
# can approve, reject, cancel, create for user / employee

insert into role_permissions (role_id, permission_id) values (3,1);
insert into role_permissions (role_id, permission_id) values (3,2);
insert into role_permissions (role_id, permission_id) values (3,3);
insert into role_permissions (role_id, permission_id) values (3,4);
insert into role_permissions (role_id, permission_id) values (3,6);

# user - me (user page), can edit, read

insert into role_permissions (role_id, permission_id) values (1,7);
insert into role_permissions (role_id, permission_id) values (1,8);

# admin - me - superuser

insert into role_permissions (role_id, permission_id) values (2,7);
insert into role_permissions (role_id, permission_id) values (2,8);
insert into role_permissions (role_id, permission_id) values (2,9);

# approver - me - can view, edit

insert into role_permissions (role_id, permission_id) values (3,7);
insert into role_permissions (role_id, permission_id) values (3,8);

CREATE TABLE IF NOT EXISTS absence_types (
    absence_type_id INT PRIMARY KEY,
    type_name VARCHAR(40) NOT NULL,
    code VARCHAR(40) NOT NULL UNIQUE,
    is_paid BOOLEAN NOT NULL,
    UNIQUE(type_name)
);
INSERT INTO absence_types (
	absence_type_id,
    type_name,
    code,
    is_paid
) VALUES (
	1,
    'Vacation',
    'VACATION',
	true
    );

INSERT INTO absence_types (
	absence_type_id,
    type_name,
    code,
    is_paid
) VALUES(
	2,
	'Parental Leave',
	'PARENTAL',
	true);

INSERT INTO absence_types (
	absence_type_id,
    type_name,
    code,
    is_paid
) VALUES(
	3,
	'Sick Leave',
	'SICK',
	true);
INSERT INTO absence_types (
	absence_type_id,
    type_name,
    code,
    is_paid
) VALUES(
	4,
	'Training Leave',
	'TRAINING',
	true);

insert
	into
	absence_types (
	absence_type_id,
	type_name,
	code,
	is_paid
)
values(5, 'Disability Leave', 'DISABILITY', true)

INSERT INTO absence_types (
	absence_type_id,
    type_name,
    code,
    is_paid
) VALUES(
	6,
	'Personal Leave',
	'PERSONAL',
	true);


# token refresh
create table refresh_tokens(
	id UUID primary key,
	user_id BIGINT not null references users(id) on delete cascade,
	token_hash text not null,
	expires_at timestamp not null,
	created_at timestamp not null default now(),
	revoked boolean not null default false,
	user_agent text,
	ip_address text
);

create index idx_refresh_user on refresh_tokens(user_id)

# get all user premissions

SELECT DISTINCT p.id, p.name
FROM permissions p
WHERE p.id IN (
    SELECT rp.permission_id
    FROM role_permissions rp
    WHERE rp.role_id = (
        SELECT r.id
        FROM roles r
        WHERE r.name = 'user'
    )
)
ORDER BY p.id;

create table if not exists countries(
	country_id int primary key not null,
	country_name varchar(40) not null
);

create table if not exists departmants(
	department_id bigint primary key not null,
	department_name varchar(30) not null,
	manager_id int
);

create table if not exists jobs(
	job_id bigint primary key not null,
	job_title varchar(35) not null,
	min_salary float,
	max_salary float
);

create table job_history(
	employee_id bigint primary key not null,
	start_date date not null,
	end_date date not null,
	job_id bigint not null,
	department_id bigint references departmants(department_id)
);

CREATE TABLE bank_accounts (
    bank_account_id BIGSERIAL PRIMARY KEY,
    employee_id BIGINT NOT NULL REFERENCES employees(employee_id) ON DELETE CASCADE,
    bank_name VARCHAR(50) NOT NULL,
    account_number VARCHAR(50) NOT NULL,
    is_primary BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE employee_social_data (
    social_id BIGSERIAL PRIMARY KEY,
    employee_id BIGINT UNIQUE NOT NULL REFERENCES employees(employee_id) ON DELETE CASCADE,
    jmbg VARCHAR(13) UNIQUE NOT NULL,
    tax_id VARCHAR(30),
    health_insurance_number VARCHAR(30)
);

CREATE TABLE education_levels (
    education_level_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

CREATE TABLE employee_education (
    education_id BIGSERIAL PRIMARY KEY,
    employee_id BIGINT NOT NULL REFERENCES employees(employee_id) ON DELETE CASCADE,
    education_level_id INT REFERENCES education_levels(education_level_id),
    institution_name VARCHAR(200),
    field_of_study VARCHAR(150),
    start_year INT,
    end_year INT
);

CREATE TABLE employee_children (
    child_id BIGSERIAL PRIMARY KEY,
    employee_id BIGINT NOT NULL REFERENCES employees(employee_id) ON DELETE CASCADE,
    first_name VARCHAR(50),
    birth_date DATE NOT NULL
);
CREATE TABLE employees(
    employee_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT UNIQUE NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    phone_number VARCHAR(20),
    hire_date DATE NOT NULL,
    job_id BIGINT NOT NULL REFERENCES jobs(job_id),
    department_id BIGINT NOT NULL REFERENCES departments(department_id),
    manager_id BIGINT REFERENCES employees(employee_id),
    salary NUMERIC(12,2) NOT null,
);

create table if not exists locations(
	location_id bigint not null primary key,
	street_address varchar(40),
	postal_code varchar(12),
	city varchar(30) not null,
	country_id int  references countries(country_id)
)
create table if not exists departmants(
	department_id bigint primary key not null,
	deoartment_name varchar(30) not null,
	manager_id int,
	location_id int references locations(location_id)
)

CREATE TYPE absence_status AS ENUM('PENDING','APPROVED', 'REJECTED', 'CANCELLED');


create table if not exists absence_types_history (
    history_id BIGSERIAL PRIMARY KEY,
    absence_type_id BIGINT NOT NULL,
    absence_status absence_status not null,
    valid_from TIMESTAMP NOT NULL,
    valid_to TIMESTAMP,
    changed_at TIMESTAMP NOT NULL DEFAULT now(),
    approved_by bigint references users(id),
    approved_at timestamp
);

CREATE TYPE action_audit_log AS ENUM (
    'INSERT',
    'UPDATE',
    'DELETE',
    'LINK',
    'UNLINK',
    'APPROVE',
    'REJECT'
);

CREATE TABLE audit_log (
    audit_id BIGSERIAL PRIMARY KEY,
    entity_name VARCHAR(50) NOT NULL,
    entity_id BIGINT NOT NULL,
    action action_audit_log not null,
    old_data JSONB,
    new_data JSONB,
    changed_by BIGINT REFERENCES users(id),
    changed_at TIMESTAMP NOT NULL DEFAULT now()
);
