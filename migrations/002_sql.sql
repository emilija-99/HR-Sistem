CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    first_name VARCHAR(20) NOT NULL,
    last_name VARCHAR(20) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
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
